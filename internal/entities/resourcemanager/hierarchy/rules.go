package hierarchy

func EnsureParentActive(exists bool, deleted bool) error {
	switch {
	case !exists:
		return ErrParentNotFound
	case deleted:
		return ErrParentDeleted
	default:
		return nil
	}
}

func EnsureDeleteAllowed(hasActiveChildren bool) error {
	if hasActiveChildren {
		return ErrDeleteBlocked
	}
	return nil
}

func EnsureUndeleteAllowed(parentExists bool, parentDeleted bool) error {
	switch {
	case !parentExists:
		return ErrUndeleteParentInvalid
	case parentDeleted:
		return ErrUndeleteParentInvalid
	default:
		return nil
	}
}
