package domain

import "errors"

var (
	ErrMemberNotFound   = errors.New("membre non trouvé")
	ErrNoteNotFound     = errors.New("note de renseignement non trouvée")
	ErrSightingNotFound = errors.New("observation non trouvée")
	ErrLinkNotFound     = errors.New("lien non trouvé")
	ErrInvalidID        = errors.New("ID invalide")
	ErrDuplicateChefID  = errors.New("national_chef_id déjà existant")
)
