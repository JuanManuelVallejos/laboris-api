package usecase

import "github.com/laboris/laboris-api/internal/domain"

// UserSyncUseCase reacciona a eventos de Clerk para mantener la app en línea
// con acciones que el usuario hace directamente en Clerk (borrar cuenta,
// cambiar foto de perfil).
type UserSyncUseCase struct {
	users domain.UserRepository
}

func NewUserSyncUseCase(users domain.UserRepository) *UserSyncUseCase {
	return &UserSyncUseCase{users: users}
}

// SoftDelete marca al usuario como eliminado sin borrar el registro: el
// admin sigue viéndolo en su dashboard, y no dispara el ON DELETE CASCADE
// que borraría al profesional asociado.
func (uc *UserSyncUseCase) SoftDelete(clerkID string) error {
	return uc.users.SoftDeleteByClerkID(clerkID)
}

func (uc *UserSyncUseCase) UpdateAvatarURL(clerkID, avatarURL string) error {
	return uc.users.UpdateAvatarURLByClerkID(clerkID, avatarURL)
}
