package streamer

import "context"

func WatchStatusChanges(ctx context.Context, sub *Sub) <-chan *StatusChanges {
	out := make(chan *StatusChanges, 1)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case changes, ok := <-sub.ChangeCh:
				if !ok {
					return
				}

				if stch, ok := changes.(*StatusChanges); ok {
					out <- stch
				}
			}
		}
	}()

	return out
}
