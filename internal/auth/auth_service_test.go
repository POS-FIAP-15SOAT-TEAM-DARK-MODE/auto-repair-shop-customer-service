package auth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Login(t *testing.T) {
	validDomainUser := domain.NewUser(
		"123",
		"John Doe",
		"test@example.com",
		"Senha@123",
	)

	tests := []struct {
		name          string
		loggedUser    adapters.LoginRequest
		mockRepo      func(t *testing.T) *mocks.AuthRepository
		expectedError error
	}{
		{
			name: "Should fail gracefully when email is empty",
			loggedUser: adapters.LoginRequest{
				Email:    "",
				Password: "Valid@Password",
			},
			expectedError: errors.Join(domain.ErrEmptyUserEmail),
		},
		{
			name: "Should fail gracefully when password is empty",
			loggedUser: adapters.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expectedError: errors.Join(domain.ErrEmptyUserPassword),
		},
		{
			name: "Should fail gracefully when get by email fails",
			loggedUser: adapters.LoginRequest{
				Email:    validDomainUser.Email,
				Password: validDomainUser.Password,
			},
			mockRepo: func(t *testing.T) *mocks.AuthRepository {
				mockRepo := mocks.NewAuthRepository(t)
				mockRepo.EXPECT().GetByEmail(mock.Anything, validDomainUser.Email).
					Return(domain.User{}, app.ErrDataConflict)
				return mockRepo
			},
			expectedError: app.ErrDataConflict,
		},
		{
			name: "Should fail gracefully when password does not match",
			loggedUser: adapters.LoginRequest{
				Email:    validDomainUser.Email,
				Password: "Not@Correct1",
			},
			mockRepo: func(t *testing.T) *mocks.AuthRepository {
				mockRepo := mocks.NewAuthRepository(t)
				responseUser := domain.NewUser(
					"123",
					"John Doe",
					"test@example.com",
					"Senha@123",
				)
				_ = responseUser.HashPassword(context.Background())
				mockRepo.EXPECT().GetByEmail(mock.Anything, validDomainUser.Email).Return(*responseUser, nil)
				return mockRepo
			},
			expectedError: domain.ErrInvalidUserCredentials,
		},
		{
			name: "Should fail gracefully when get roles by id fails",
			loggedUser: adapters.LoginRequest{
				Email:    validDomainUser.Email,
				Password: validDomainUser.Password,
			},
			mockRepo: func(t *testing.T) *mocks.AuthRepository {
				mockRepo := mocks.NewAuthRepository(t)
				responseUser := domain.NewUser(
					"123",
					"John Doe",
					"test@example.com",
					"Senha@123",
				)
				_ = responseUser.HashPassword(context.Background())
				mockRepo.EXPECT().GetByEmail(mock.Anything, validDomainUser.Email).Return(*responseUser, nil)
				mockRepo.EXPECT().GetUserSession(mock.Anything, validDomainUser.ID).Return(adapters.LoginResponse{}, app.ErrDataViolation)
				return mockRepo
			},
			expectedError: app.ErrDataViolation,
		},
		{
			name: "Should return the token successfully",
			loggedUser: adapters.LoginRequest{
				Email:    validDomainUser.Email,
				Password: validDomainUser.Password,
			},
			mockRepo: func(t *testing.T) *mocks.AuthRepository {
				mockRepo := mocks.NewAuthRepository(t)
				responseUser := domain.NewUser(
					"123",
					"John Doe",
					"test@example.com",
					"Senha@123",
				)
				_ = responseUser.HashPassword(context.Background())
				mockRepo.EXPECT().GetByEmail(mock.Anything, validDomainUser.Email).Return(*responseUser, nil)
				mockRepo.EXPECT().GetUserSession(mock.Anything, validDomainUser.ID).Return(adapters.LoginResponse{}, nil)
				return mockRepo
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockRepo *mocks.AuthRepository
			if tt.mockRepo != nil {
				mockRepo = tt.mockRepo(t)
			}

			service := auth.NewService(nil, mockRepo, nil)
			_, err := service.Login(context.Background(), tt.loggedUser)
			assert.Equal(t, err, tt.expectedError)
		})
	}
}

func TestService_Register(t *testing.T) {
	validCreateUser := adapters.CreateUserRequest{
		Name:            "valid user",
		Email:           "valid@email.com",
		Password:        "Valid@Password",
		ConfirmPassword: "Valid@Password",
	}
	tests := []struct {
		name          string
		data          adapters.CreateUserRequest
		prepareMocks  func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository)
		expectedError error
	}{
		{
			name: "Should fail gracefully when password don't match confirmation",
			data: adapters.CreateUserRequest{
				Password:        "password",
				ConfirmPassword: "password1",
			},
			expectedError: domain.ErrUserPasswordDontMatch,
		},
		{
			name:          "Should fail gracefully when params are invalid",
			data:          adapters.CreateUserRequest{},
			expectedError: errors.Join(domain.ErrEmptyUserName, domain.ErrEmptyUserEmail, domain.ErrEmptyUserPassword),
		},
		{
			name: "Should fail gracefully when hashing password fails",
			data: adapters.CreateUserRequest{
				Name:            "valid user",
				Email:           "valid@email.com",
				Password:        fmt.Sprintf("Valid@1%v", string(make([]byte, 73))),
				ConfirmPassword: fmt.Sprintf("Valid@1%v", string(make([]byte, 73))),
			},
			expectedError: domain.ErrUserPasswordTooLong,
		},
		{
			name: "Should fail gracefully when transaction fails",
			data: validCreateUser,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(app.ErrInfraConflict)
			},
			expectedError: app.ErrInfraConflict,
		},
		{
			name: "Should fail gracefully when save repository fails",
			data: validCreateUser,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
					for _, step := range steps {
						if err := step(ctx); err != nil {
							return err
						}
					}
					return nil
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.User")).Return(app.ErrDataConflict)
			},
			expectedError: app.ErrDataConflict,
		},
		{
			name: "Should fail gracefully when update role repository fails",
			data: validCreateUser,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
					for _, step := range steps {
						if err := step(ctx); err != nil {
							return err
						}
					}
					return nil
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
				repo.EXPECT().SaveUserRole(mock.Anything, "test", domain.CUSTOMER).Return(app.ErrDataConflict)
			},
			expectedError: app.ErrDataConflict,
		},
		{
			name: "Should register successfully",
			data: validCreateUser,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
					for _, step := range steps {
						if err := step(ctx); err != nil {
							return err
						}
					}
					return nil
				})
				repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
				repo.EXPECT().SaveUserRole(mock.Anything, "test", domain.CUSTOMER).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			repo := mocks.NewAuthRepository(t)
			if tt.prepareMocks != nil {
				tt.prepareMocks(t, executor, repo)
			}

			svc := auth.NewService(executor, repo, id.NewNoopIDGenerator("test"))
			_, err := svc.Register(context.Background(), tt.data)
			assert.Equal(t, err, tt.expectedError)
		})
	}
}

func TestService_ChangeRole(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		role          domain.Role
		prepareMocks  func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository)
		expectedError error
	}{
		{
			name:          "Should fail gracefully when role is invalid",
			id:            "user-1",
			role:          domain.Role("INVALID"),
			expectedError: domain.ErrInvalidUserCredentials,
		},
		{
			name: "Should fail gracefully when repository fails",
			id:   "user-1",
			role: domain.MECHANIC,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
					for _, step := range steps {
						if err := step(ctx); err != nil {
							return err
						}
					}
					return nil
				})
				repo.EXPECT().SaveUserRole(mock.Anything, "user-1", domain.MECHANIC).Return(app.ErrDataConflict)
			},
			expectedError: app.ErrDataConflict,
		},
		{
			name: "Should fail gracefully when transaction fails",
			id:   "user-1",
			role: domain.MECHANIC,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(app.ErrInfraConflict)
			},
			expectedError: app.ErrInfraConflict,
		},
		{
			name: "Should update role successfully",
			id:   "user-1",
			role: domain.MECHANIC,
			prepareMocks: func(t *testing.T, executor *uowmocks.Executor, repo *mocks.AuthRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
					for _, step := range steps {
						if err := step(ctx); err != nil {
							return err
						}
					}
					return nil
				})
				repo.EXPECT().SaveUserRole(mock.Anything, "user-1", domain.MECHANIC).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			repo := mocks.NewAuthRepository(t)
			if tt.prepareMocks != nil {
				tt.prepareMocks(t, executor, repo)
			}

			svc := auth.NewService(executor, repo, nil)
			err := svc.ChangeRole(context.Background(), tt.id, adapters.ChangeRoleRequest{Role: tt.role})
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}
