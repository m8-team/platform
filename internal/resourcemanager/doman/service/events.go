package service

type EventType string

const (
	EventServiceCreated EventType = "resourcemanager.service.create"
	EventServiceUpdated EventType = "resourcemanager.service.update"
	EventServiceDeleted EventType = "resourcemanager.service.delete"
)
