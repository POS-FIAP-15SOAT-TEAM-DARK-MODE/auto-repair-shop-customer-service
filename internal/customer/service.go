package customer

import (
	"context"

	authAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	authInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
)

type customerService struct {
	uow          uow.Executor
	customerRepo interfaces.CustomerRepository
	authRepo     authInterfaces.AuthRepository
}

func NewService(u uow.Executor, customerRepo interfaces.CustomerRepository, authRepo authInterfaces.AuthRepository) interfaces.CustomerService {
	return &customerService{uow: u, customerRepo: customerRepo, authRepo: authRepo}
}

func (s *customerService) Create(ctx context.Context, req adapters.CreateCustomerRequest) (adapters.CustomerResponse, error) {
	customer, err := domain.NewCustomer(req.Name, req.Email, domain.CustomerType(req.Type), req.Document, req.CompanyName, req.Phone)
	if err != nil {
		return adapters.CustomerResponse{}, err
	}

	// Default password is the sanitized document (CPF or CNPJ)
	if customer.Type == domain.IndividualCustomerType {
		customer.User.Password = customer.CPF
	} else {
		customer.User.Password = customer.CNPJ
	}

	if err = customer.User.HashPassword(ctx); err != nil {
		return adapters.CustomerResponse{}, err
	}

	if err = s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.authRepo.Save(txCtx, customer.User)
		},
		func(txCtx context.Context) error {
			return s.authRepo.SaveUserRole(txCtx, customer.User.ID, authDomain.CUSTOMER)
		},
		func(txCtx context.Context) error {
			return s.customerRepo.Save(txCtx, customer)
		},
	); err != nil {
		return adapters.CustomerResponse{}, err
	}

	return domainToResponse(*customer), nil
}

func (s *customerService) GetByID(ctx context.Context, id string) (adapters.CustomerResponse, error) {
	if id == "" {
		return adapters.CustomerResponse{}, domain.ErrInvalidCustomerId
	}

	customers, err := s.customerRepo.List(ctx, adapters.ListCustomerParams{ID: id})
	if err != nil {
		return adapters.CustomerResponse{}, err
	}
	if len(customers) == 0 {
		return adapters.CustomerResponse{}, domain.ErrCustomerNotFound
	}

	return domainToResponse(customers[0]), nil
}

func (s *customerService) GetByDocument(ctx context.Context, rawDocument string) (adapters.CustomerResponse, error) {
	_, document, err := domain.ParseDocument(rawDocument)
	if err != nil {
		return adapters.CustomerResponse{}, err
	}

	customers, err := s.customerRepo.List(ctx, adapters.ListCustomerParams{Document: document})
	if err != nil {
		return adapters.CustomerResponse{}, err
	}
	if len(customers) == 0 {
		return adapters.CustomerResponse{}, domain.ErrCustomerNotFound
	}

	return domainToResponse(customers[0]), nil
}

func (s *customerService) Update(ctx context.Context, id string, req adapters.UpdateCustomerRequest) (adapters.CustomerResponse, error) {
	customers, err := s.customerRepo.List(ctx, adapters.ListCustomerParams{ID: id})
	if err != nil {
		return adapters.CustomerResponse{}, err
	}
	if len(customers) == 0 {
		return adapters.CustomerResponse{}, domain.ErrCustomerNotFound
	}
	c := customers[0]

	if req.Name != "" {
		c.User.Name = req.Name
	}
	if req.Email != "" {
		c.User.Email = req.Email
	}
	if req.Phone != "" {
		c.Phone = req.Phone
	}

	if err = s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.authRepo.Update(txCtx, c.UserID, c.User.Name, c.User.Email)
		},
		func(txCtx context.Context) error {
			return s.customerRepo.Save(txCtx, &c)
		},
	); err != nil {
		return adapters.CustomerResponse{}, err
	}

	return domainToResponse(c), nil
}

func (s *customerService) Delete(ctx context.Context, id string) error {
	customers, err := s.customerRepo.List(ctx, adapters.ListCustomerParams{ID: id})
	if err != nil {
		return err
	}
	if len(customers) == 0 {
		return nil
	}
	c := customers[0]

	return s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.customerRepo.Delete(txCtx, id)
		},
		func(txCtx context.Context) error {
			return s.authRepo.Delete(txCtx, c.UserID)
		},
	)
}

func domainToResponse(c domain.Customer) adapters.CustomerResponse {
	resp := adapters.CustomerResponse{
		ID:          c.ID,
		Type:        string(c.Type),
		CompanyName: c.CompanyName,
		Phone:       c.Phone,
	}

	if c.Type == domain.IndividualCustomerType {
		resp.Document = c.CPF
	} else {
		resp.Document = c.CNPJ
	}

	if c.User != nil {
		resp.UserResponse = authAdapters.UserResponse{
			ID:    c.User.ID,
			Name:  c.User.Name,
			Email: c.User.Email,
		}
	}

	return resp
}
