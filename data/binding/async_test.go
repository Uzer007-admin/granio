package binding

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAsyncListRejectsStaleProviderResult(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	list := NewAsyncList(ctx, func(context.Context) ([]string, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-release // intentionally ignore cancellation to test late completion
			return []string{"stale"}, nil
		}
		return []string{"current"}, nil
	}, func(a, b string) bool { return a == b })
	defer list.Close()
	<-started
	list.mu.Lock()
	first := list.current
	list.mu.Unlock()
	require.NoError(t, list.Reload())
	require.NoError(t, list.Wait(ctx))
	close(release)
	select {
	case <-first.done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	value, err := list.Get()
	require.NoError(t, err)
	require.Equal(t, []string{"current"}, value)
	require.ErrorIs(t, first.err, context.Canceled)
	value[0] = "caller copy"
	value, _ = list.Get()
	require.Equal(t, []string{"current"}, value)
}

func TestAsyncListFailureRetainsDataAndCloseCancels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	failed := errors.New("provider failed")
	var calls atomic.Int32
	list := NewAsyncList(ctx, func(ctx context.Context) ([]int, error) {
		switch calls.Add(1) {
		case 1:
			return []int{1, 2}, nil
		case 2:
			return nil, failed
		default:
			<-ctx.Done()
			return nil, ctx.Err()
		}
	}, func(a, b int) bool { return a == b })
	defer list.Close()
	require.NoError(t, list.Wait(ctx))
	require.NoError(t, list.Reload())
	require.ErrorIs(t, list.Wait(ctx), failed)
	require.ErrorIs(t, list.LoadError(), failed)
	values, _ := list.Get()
	require.Equal(t, []int{1, 2}, values)
	require.NoError(t, list.Reload())
	require.True(t, list.Loading())
	list.Close()
	require.ErrorIs(t, list.Wait(ctx), context.Canceled)
	require.False(t, list.Loading())
	require.Error(t, list.Reload())
}

func TestAsyncTreeCommitsHierarchySnapshot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tree := NewAsyncTree(ctx, func(context.Context) (TreeSnapshot[string], error) {
		return TreeSnapshot[string]{Children: map[string][]string{"": {"branch"}, "branch": {"leaf"}},
			Values: map[string]string{"": "root", "branch": "parent", "leaf": "child"}}, nil
	}, func(a, b string) bool { return a == b })
	defer tree.Close()
	require.NoError(t, tree.Wait(ctx))
	require.Equal(t, []string{"branch"}, tree.ChildIDs(""))
	require.Equal(t, []string{"leaf"}, tree.ChildIDs("branch"))
	value, err := tree.GetValue("leaf")
	require.NoError(t, err)
	require.Equal(t, "child", value)
}
