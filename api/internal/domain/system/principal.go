package system

import "errors"

var (
	ErrOwnerRequired = errors.New("ação restrita exclusivamente ao dono com autenticação 2FA (AAL2)")
)

// Principal representa a identidade que executa uma ação no sistema.
type Principal interface {
	ActorID() string
	IsOwner() bool
}

// OwnerPrincipal representa o único dono do sistema autenticado com 2FA.
type OwnerPrincipal struct {
	ID  string
	AAL string
}

func (o OwnerPrincipal) ActorID() string {
	return o.ID
}

func (o OwnerPrincipal) IsOwner() bool {
	return true
}

// NewOwnerPrincipal constrói e valida a identidade do dono exigindo AAL2.
func NewOwnerPrincipal(id, aal string) (OwnerPrincipal, error) {
	if id == "" {
		return OwnerPrincipal{}, errors.New("id do dono é obrigatório")
	}
	if aal != "aal2" {
		return OwnerPrincipal{}, ErrOwnerRequired
	}
	return OwnerPrincipal{
		ID:  id,
		AAL: aal,
	}, nil
}

// AgentPrincipal representa um agente de IA do sistema (futuro, desligado por padrão).
type AgentPrincipal struct {
	ID              string
	Role            string
	Active          bool
	SpendLimitCents int64
}

func (a AgentPrincipal) ActorID() string {
	return a.ID
}

func (a AgentPrincipal) IsOwner() bool {
	return false
}
