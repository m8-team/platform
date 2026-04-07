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

func EnsureUndeleteAllowed(exists bool, deleted bool) error {
	switch {
	case !exists:
		return ErrUndeleteParentInvalid
	case deleted:
		return ErrUndeleteParentInvalid
	default:
		return nil
	}
}
