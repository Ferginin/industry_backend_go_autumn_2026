package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	lim := &Limiter{
		clock:  clock,
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(burst),
	}
	// проверка, чтобы без паники
	if lim.clock != nil {
		lim.last = clock.Now()
	}
	return lim
}

func (l *Limiter) AllowN(n int) bool {
	// стандартная блокировка
	l.mu.Lock()
	defer l.mu.Unlock()

	// все проверки по условиям
	if l.burst <= 0 || l.clock == nil {
		return false
	}
	if n <= 0 || n > l.burst {
		return false
	}

	// один раз берем  текущее время, дальше работаем только с ним
	now := l.clock.Now()

	// вычисляем сколько нужно добавить токенов,
	// если rate <= 0, то просто ничего не добавится
	addTokens := l.rate * now.Sub(l.last).Seconds()
	if addTokens > 0 {
		l.tokens += addTokens
		// ловим переполнение
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
		// и меняем время
		l.last = now
	}

	// не отнимаем если n больше текущих токенов
	if l.tokens < float64(n) {
		return false
	} else {
		l.tokens -= float64(n)
	}

	return true
}
