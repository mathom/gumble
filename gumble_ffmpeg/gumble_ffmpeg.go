package gumble_ffmpeg

import (
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/layeh/gumble/gumble"
)

const (
	DefaultCommand = "ffmpeg"
)

type Stream struct {
	Command string
	Volume  float32
	ElapsedTime float64

	playLock sync.Mutex
	playStart time.Time

	client *gumble.Client
	cmd    *exec.Cmd
	pipe   io.ReadCloser

	stop          chan bool
	stopWaitGroup sync.WaitGroup
}

func New(client *gumble.Client) (*Stream, error) {
	stream := &Stream{
		client:  client,
		Volume:  1.0,
		Command: DefaultCommand,
		stop:    make(chan bool),
	}
	return stream, nil
}

func (s *Stream) PlayExec(name string, args []string, callbacks ...func()) error {
	cmd := exec.Command(name, args...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	cb := func() {
		cmd.Process.Kill()
		cmd.Wait()
	}
	callbacks = append(callbacks, cb)
	err = s.play("-", pipe, 0.0, callbacks...)
	if err != nil {
		cb()
		return err
	}
	return nil
}

func (s *Stream) PlayReader(in io.Reader, callbacks ...func()) error {
	return s.play("-", in, 0.0, callbacks...)
}

func (s *Stream) Play(file string, startSeconds float64, callbacks ...func()) error {
	return s.play(file, nil, startSeconds, callbacks...)
}

func (s *Stream) play(file string, in io.Reader, startSeconds float64, callbacks ...func()) error {
	s.playLock.Lock()
	defer s.playLock.Unlock()

	if s.IsPlaying() {
		return errors.New("already playing")
	}
	s.playStart = time.Now().Add(time.Duration(-startSeconds)*time.Second)
	s.ElapsedTime = 0.0
    
	cmd_args := []string{"-i", file, "-ac", "1", "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-"}
	if (startSeconds > 0) {
		cmd_args = append([]string{"-ss", strconv.FormatFloat(startSeconds, 'f', -1, 32)}, cmd_args...)
	}

	cmd := exec.Command(s.Command, cmd_args...)

	if pipe, err := cmd.StdoutPipe(); err != nil {
		return err
	} else {
		s.pipe = pipe
	}
	if in != nil {
		cmd.Stdin = in
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.stopWaitGroup.Add(1)
	s.cmd = cmd
	go s.sourceRoutine(callbacks)
	return nil
}

func (s *Stream) IsPlaying() bool {
	return s.cmd != nil
}

func (s *Stream) Wait() {
	s.stopWaitGroup.Wait()
}

func (s *Stream) Stop() error {
	if !s.IsPlaying() {
		return errors.New("nothing playing")
	}

	s.stop <- true
	s.stopWaitGroup.Wait()
	return nil
}

func (s *Stream) sourceRoutine(callbacks []func()) {
	interval := s.client.Config.GetAudioInterval()
	frameSize := s.client.Config.GetAudioFrameSize()

	ticker := time.NewTicker(interval)

	defer func() {
		ticker.Stop()
		s.cmd.Process.Kill()
		s.cmd.Wait()
		s.cmd = nil
		s.stopWaitGroup.Done()
		for _, callback := range callbacks {
			callback()
		}
	}()

	int16Buffer := make([]int16, frameSize)
	byteBuffer := make([]byte, frameSize*2)

	for {
		select {
		case <-s.stop:
			s.ElapsedTime = time.Since(s.playStart).Seconds()
			return
		case <-ticker.C:
			if _, err := io.ReadFull(s.pipe, byteBuffer); err != nil {
				return
			}
			for i := range int16Buffer {
				float := float32(int16(binary.LittleEndian.Uint16(byteBuffer[i*2 : (i+1)*2])))
				int16Buffer[i] = int16(s.Volume * float)
			}
			s.ElapsedTime = time.Since(s.playStart).Seconds()
			s.client.Send(gumble.AudioBuffer(int16Buffer))
		}
	}
}
