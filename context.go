package errx

import "context"

// errxCtxKey is the private context key for errx metadata.
type errxCtxKey struct{}

// WithMetaContext stores key-value metadata in the context for later attachment to errors
// via [Error.WithMetaFromContext]. It accepts alternating key-value pairs where each key
// should be a string. Non-string keys are silently skipped, and a trailing key
// with no value is silently dropped.
//
// Each call copies the parent context's metadata into a new map, then applies the
// provided key-value pairs on top (last-write-wins). The parent's map is never
// mutated, so concurrent goroutines that derive from the same parent context each
// get an independent snapshot with the correct metadata for their scope.
//
// Concurrency example:
//
//	ctx = errx.WithMetaContext(ctx, "request_id", reqID)
//
//	for _, postID := range postIDs {
//	    group.Go(func() error {
//	        ctx := errx.WithMetaContext(ctx, "post_id", postID)
//	        // Each goroutine gets its own metadata snapshot.
//	        // Errors created here carry post_id specific to this goroutine.
//	        return errx.NewInternal("failed").WithMetaFromContext(ctx)
//	    })
//	}
//
//	// Whatever error the errgroup returns will have the correct post_id
//	// for the goroutine that failed — no cross-contamination.
//
// Basic usage:
//
//	ctx = errx.WithMetaContext(ctx, "user_id", 123, "action", "delete")
//	err := errx.New(errx.CodeNotFound, "user not found").WithMetaFromContext(ctx)
func WithMetaContext(ctx context.Context, keyvals ...any) context.Context {
	existing := getCtxMeta(ctx)

	merged := make(map[string]any, len(existing)+len(keyvals)/2)
	for k, v := range existing {
		merged[k] = v
	}

	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		merged[key] = keyvals[i+1]
	}

	return context.WithValue(ctx, errxCtxKey{}, merged)
}

// getCtxMeta retrieves errx metadata from the context.
// Returns nil if no metadata is stored.
func getCtxMeta(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	meta, ok := ctx.Value(errxCtxKey{}).(map[string]any)
	if !ok {
		return nil
	}
	return meta
}
