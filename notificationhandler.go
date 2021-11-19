package main

import "strings"

var globalNotificationHandler func(NotificationEvent) = nil

type NotificationEvent struct {
	Type    NotificationType
	Message string
}

func NotifyError(message string) {
	if globalNotificationHandler == nil {
		return
	}

	var sb strings.Builder
	sb.WriteString("[ERROR] ")
	sb.WriteString(message)
	globalNotificationHandler(NotificationEvent{Type: NotificationError, Message: sb.String()})
}

func NotifyInfo(message string) {
	if globalNotificationHandler == nil {
		return
	}

	var sb strings.Builder
	sb.WriteString("[INFO] ")
	sb.WriteString(message)
	globalNotificationHandler(NotificationEvent{Type: NotificationInfo, Message: sb.String()})
}
