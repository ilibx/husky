package ldap

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/husky/husky/internal/config"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/logger"
	ldapv3 "github.com/go-ldap/ldap/v3"
	"golang.org/x/crypto/bcrypt"
)

type SyncResult struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

type Service struct {
	userRepo *repository.UserRepository
	log      *logger.Logger
}

func NewService(userRepo *repository.UserRepository, log *logger.Logger) *Service {
	return &Service{userRepo: userRepo, log: log}
}

func (s *Service) SyncUsers(ctx context.Context) (*SyncResult, error) {
	cfg := config.Conf
	if cfg.LDAPHost == "" {
		return nil, fmt.Errorf("LDAP not configured: LDAP_HOST is empty")
	}

	var fieldMap map[string]string
	if err := json.Unmarshal([]byte(cfg.LDAPFieldMap), &fieldMap); err != nil {
		return nil, fmt.Errorf("invalid LDAP_FIELD_MAP: %w", err)
	}

	conn, err := ldapv3.Dial("tcp", fmt.Sprintf("%s:%d", cfg.LDAPHost, cfg.LDAPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer conn.Close()

	if err := conn.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		s.log.Warn("LDAP STARTTLS failed, trying plain connection", "error", err)
	}

	if cfg.LDAPBindDN != "" {
		if err := conn.Bind(cfg.LDAPBindDN, cfg.LDAPPassword); err != nil {
			return nil, fmt.Errorf("LDAP bind failed: %w", err)
		}
	}

	searchReq := ldapv3.NewSearchRequest(
		cfg.LDAPBaseDN,
		ldapv3.ScopeWholeSubtree,
		ldapv3.NeverDerefAliases,
		0, 0, false,
		cfg.LDAPFilter,
		[]string{}, // all attributes
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	res := &SyncResult{Total: len(result.Entries)}

	for _, entry := range result.Entries {
		if err := s.syncEntry(ctx, entry, fieldMap); err != nil {
			s.log.Error("LDAP sync entry failed", "error", err, "dn", entry.DN)
			res.Errors++
			continue
		}
		res.Created++
	}

	s.log.Info("LDAP sync completed", "total", res.Total, "created", res.Created, "errors", res.Errors)
	return res, nil
}

func (s *Service) syncEntry(ctx context.Context, entry *ldapv3.Entry, fieldMap map[string]string) error {
	username := entry.GetAttributeValue("cn")
	email := entry.GetAttributeValue("mail")
	if email == "" {
		s.log.Debug("LDAP entry skipped: no email", "dn", entry.DN)
		return nil
	}

	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil
	}

	phone := entry.GetAttributeValue("telephoneNumber")
	title := entry.GetAttributeValue("title")
	department := entry.GetAttributeValue("department")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("changeme123"), bcrypt.DefaultCost)

	user := &model.User{
		Email:      email,
		Username:   username,
		Password:   string(hashedPassword),
		Phone:      phone,
		Title:      title,
		Department: department,
		Role:       "user",
		Status:     1,
	}

	return s.userRepo.Create(ctx, user)
}
