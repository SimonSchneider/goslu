package syncu

type Broadcaster[T any] struct {
	stopCh        chan struct{}
	publishCh     chan T
	subCh         chan chan T
	unsubCh       chan chan T
	OnSubscribe   func(chan T)
	OnUnsubscribe func(chan T)
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		stopCh:    make(chan struct{}),
		publishCh: make(chan T, 1),
		subCh:     make(chan chan T, 1),
		unsubCh:   make(chan chan T, 1),
	}
}

func (b *Broadcaster[T]) Start() {
	go b.Run()
}

func (b *Broadcaster[T]) Run() {
	subs := map[chan T]struct{}{}
	for {
		select {
		case <-b.stopCh:
			for msgCh := range subs {
				close(msgCh)
			}
			close(b.publishCh)
			return
		case msgCh := <-b.subCh:
			if b.OnSubscribe != nil {
				b.OnSubscribe(msgCh)
			}
			subs[msgCh] = struct{}{}
		case msgCh := <-b.unsubCh:
			if b.OnUnsubscribe != nil {
				b.OnUnsubscribe(msgCh)
			}
			delete(subs, msgCh)
			close(msgCh)
		case msg := <-b.publishCh:
			for msgCh := range subs {
				// msgCh is buffered, use non-blocking send to protect the broker:
				select {
				case msgCh <- msg:
				default:
				}
			}
		}
	}
}

func (b *Broadcaster[T]) Close() {
	select {
	case <-b.stopCh:
		return
	default:
	}
	close(b.stopCh)
}

func (b *Broadcaster[T]) Subscribe() (chan T, func()) {
	msgCh := make(chan T, 5)
	b.subCh <- msgCh
	return msgCh, func() {
		select {
		case b.unsubCh <- msgCh:
		case <-b.stopCh:
		}
	}
}

func (b *Broadcaster[T]) Publish(msg T) {
	b.publishCh <- msg
}
