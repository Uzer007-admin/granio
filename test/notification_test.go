package test

import (
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"

	"github.com/stretchr/testify/assert"
)

func TestAssertNotificationSent(t *testing.T) {
	n := granio.NewNotification("Test Title", "Some content")
	myApp := granio.CurrentApp()

	AssertNotificationSent(t, n, func() {
		granio.CurrentApp().SendNotification(n)
	})
	assert.Equal(t, myApp, granio.CurrentApp())
}

func TestAssertNotificationSent_Nil(t *testing.T) {
	AssertNotificationSent(t, nil, func() {
		// don't send anything
	})
}

func TestAssertNotificationSent_NotSent(t *testing.T) {
	tt := &testing.T{}

	AssertNotificationSent(tt, &granio.Notification{}, func() {
		// don't send anything
	})
	assert.True(t, tt.Failed(), "notification assert should fail if no notification was sent")
}

func TestAssertNotificationScheduled(t *testing.T) {
	n := granio.NewNotification("Reminder", "Take a break")
	myApp := granio.CurrentApp()

	var scheduled *granio.ScheduledNotification
	AssertNotificationScheduled(t, n, func() {
		s, err := granio.CurrentApp().ScheduleNotification(n, time.Now().Add(time.Minute))
		assert.NoError(t, err)
		scheduled = s
	})
	assert.NotNil(t, scheduled)
	assert.NotEmpty(t, scheduled.ID())
	assert.Equal(t, myApp, granio.CurrentApp())
}

func TestAssertNotificationScheduled_CancelClearsState(t *testing.T) {
	n := granio.NewNotification("Reminder", "Take a break")
	s, err := granio.CurrentApp().ScheduleNotification(n, time.Now().Add(time.Minute))
	assert.NoError(t, err)

	err = granio.CurrentApp().CancelScheduledNotification(s.ID())
	assert.NoError(t, err)
}
