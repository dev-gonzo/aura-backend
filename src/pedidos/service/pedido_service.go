package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"sistema-editorial/editora/backend/src/pedidos/entity"
)

type repository interface {
	NextCode(ctx context.Context) (string, error)
	LookupLivro(ctx context.Context, id string) (string, string, error)
	Create(ctx context.Context, input entity.PersistInput) (string, error)
	Update(ctx context.Context, input entity.PersistInput) error
	List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error)
	FindByID(ctx context.Context, id string) (entity.DetailResponse, error)
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error) {
	return s.repo.List(ctx, query)
}

func (s *Service) FindByID(ctx context.Context, id string) (entity.DetailResponse, error) {
	if strings.TrimSpace(id) == "" {
		return entity.DetailResponse{}, ValidationError{Message: "id do pedido e obrigatorio"}
	}
	return s.repo.FindByID(ctx, strings.TrimSpace(id))
}

func (s *Service) Create(ctx context.Context, request entity.CreateRequest) (string, error) {
	input, err := s.normalizePersistInput(ctx, "", request)
	if err != nil {
		return "", err
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, request entity.UpdateRequest) error {
	input, err := s.normalizePersistInput(ctx, id, request)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, input)
}

func (s *Service) normalizePersistInput(
	ctx context.Context,
	id string,
	request entity.CreateRequest,
) (entity.PersistInput, error) {
	clienteNome := strings.TrimSpace(request.ClienteNome)
	if clienteNome == "" {
		return entity.PersistInput{}, ValidationError{Message: "cliente do pedido e obrigatorio"}
	}
	if len(request.Itens) == 0 {
		return entity.PersistInput{}, ValidationError{Message: "o pedido deve ter pelo menos um item"}
	}

	status := normalizeOrderStatus(request.Status)
	if status == "" {
		status = "RASCUNHO"
	}
	canalVenda := strings.ToUpper(strings.TrimSpace(request.CanalVenda))
	if canalVenda == "" {
		canalVenda = "MANUAL"
	}
	if canalVenda != "MANUAL" && canalVenda != "ECOMMERCE" {
		return entity.PersistInput{}, ValidationError{Message: "canal de venda invalido"}
	}

	items := make([]entity.PersistItem, 0, len(request.Itens))
	subtotal := 0.0
	for _, item := range request.Itens {
		livroID := strings.TrimSpace(item.LivroID)
		if livroID == "" {
			return entity.PersistInput{}, ValidationError{Message: "livro do pedido e obrigatorio"}
		}
		if item.Quantidade <= 0 {
			return entity.PersistInput{}, ValidationError{Message: "quantidade do item deve ser maior que zero"}
		}
		if item.PrecoUnitario < 0 {
			return entity.PersistInput{}, ValidationError{Message: "preco do item nao pode ser negativo"}
		}
		titulo, autor, err := s.repo.LookupLivro(ctx, livroID)
		if err != nil {
			return entity.PersistInput{}, fmt.Errorf("erro ao localizar livro do pedido: %w", err)
		}
		itemSubtotal := float64(item.Quantidade) * item.PrecoUnitario
		subtotal += itemSubtotal
		items = append(items, entity.PersistItem{
			LivroID:       livroID,
			TituloLivro:   titulo,
			AutorNome:     autor,
			Quantidade:    item.Quantidade,
			PrecoUnitario: item.PrecoUnitario,
			Subtotal:      itemSubtotal,
		})
	}

	if request.Desconto < 0 || request.Frete < 0 {
		return entity.PersistInput{}, ValidationError{Message: "valores de frete e desconto nao podem ser negativos"}
	}

	total := subtotal - request.Desconto + request.Frete
	if total < 0 {
		total = 0
	}

	codigo := ""
	if strings.TrimSpace(id) == "" {
		nextCode, err := s.repo.NextCode(ctx)
		if err != nil {
			return entity.PersistInput{}, fmt.Errorf("erro ao gerar codigo do pedido: %w", err)
		}
		codigo = nextCode
	}

	return entity.PersistInput{
		ID:              strings.TrimSpace(id),
		Codigo:          codigo,
		CanalVenda:      canalVenda,
		Status:          status,
		ClienteNome:     clienteNome,
		ClienteEmail:    normalizeOptionalString(request.ClienteEmail),
		ClienteWhatsapp: normalizeOptionalString(request.ClienteWhatsapp),
		Subtotal:        subtotal,
		Desconto:        request.Desconto,
		Frete:           request.Frete,
		Total:           total,
		Observacao:      normalizeOptionalString(request.Observacao),
		Itens:           items,
		Entrega:         normalizeEntrega(request.Entrega),
	}, nil
}

func normalizeEntrega(request *entity.EntregaRequest) *entity.PersistEntrega {
	if request == nil {
		return nil
	}
	tipoEntrega := strings.ToUpper(strings.TrimSpace(request.TipoEntrega))
	if tipoEntrega == "" {
		tipoEntrega = "CORREIOS"
	}
	statusEntrega := strings.ToUpper(strings.TrimSpace(request.StatusEntrega))
	if statusEntrega == "" {
		statusEntrega = "PENDENTE"
	}
	destinatarioNome := strings.TrimSpace(request.DestinatarioNome)
	if destinatarioNome == "" {
		destinatarioNome = "Destinatario"
	}

	return &entity.PersistEntrega{
		TipoEntrega:           tipoEntrega,
		StatusEntrega:         statusEntrega,
		Transportadora:        normalizeOptionalString(request.Transportadora),
		CodigoRastreio:        normalizeOptionalString(request.CodigoRastreio),
		DestinatarioNome:      destinatarioNome,
		DestinatarioDocumento: normalizeOptionalString(request.DestinatarioDocumento),
		CEP:                   normalizeOptionalString(request.CEP),
		Logradouro:            normalizeOptionalString(request.Logradouro),
		Numero:                normalizeOptionalString(request.Numero),
		Complemento:           normalizeOptionalString(request.Complemento),
		Bairro:                normalizeOptionalString(request.Bairro),
		Cidade:                normalizeOptionalString(request.Cidade),
		UF:                    normalizeOptionalString(request.UF),
		PrazoPrevistoEm:       normalizeOptionalString(request.PrazoPrevistoEm),
		PostadoEm:             normalizeOptionalString(request.PostadoEm),
		EntregueEm:            normalizeOptionalString(request.EntregueEm),
		Observacao:            normalizeOptionalString(request.Observacao),
	}
}

func normalizeOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOrderStatus(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case "", "RASCUNHO", "AGUARDANDO_PAGAMENTO", "PAGO", "EM_SEPARACAO", "ENVIADO", "ENTREGUE", "CANCELADO":
		return normalized
	default:
		return ""
	}
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
