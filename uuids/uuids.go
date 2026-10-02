package uuids

import "uuid"

func UuidV4() uuid.UUID {
	return uuid.NewV4()
}

func UuidV4String() string {
	return uuid.NewV4().String()
}

func UuidV7() uuid.UUID {
	return uuid.NewV7()
}

func UuidV7String() string {
	return uuid.NewV7().String()
}
