package clock

import "time"

// Clock abstrai a medição de tempo no domínio, permitindo testes determinísticos.
type Clock interface {
	Now() time.Time
}

// SystemClock implementa Clock usando o relógio real do sistema operacional.
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

// FrozenClock implementa Clock com um horário fixo para testes de unidade.
type FrozenClock struct {
	Current time.Time
}

func (f FrozenClock) Now() time.Time {
	return f.Current
}
