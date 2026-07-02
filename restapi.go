package spworlds

import (
	"context"
	"fmt"
)

func (c *Client) ClientCard(ctx context.Context) (*ClientCard, error) {
	var out ClientCard
	if err := c.get(ctx, "card", &out); err != nil {
		return nil, fmt.Errorf("ClientCard: %w", err)
	}
	return &out, nil
}

func (c *Client) User(ctx context.Context, discordID string) (*User, error) {
	var out User
	if err := c.get(ctx, "users/"+discordID, &out); err != nil {
		return nil, fmt.Errorf("User(%s): %w", discordID, err)
	}
	return &out, nil
}

func (c *Client) UserCards(ctx context.Context, username string) ([]Card, error) {
	var out []Card
	if err := c.get(ctx, "accounts/"+username+"/cards", &out); err != nil {
		return nil, fmt.Errorf("UserCards(%s): %w", username, err)
	}
	return out, nil
}

func (c *Client) Me(ctx context.Context) (*SelfUser, error) {
	var out SelfUser
	if err := c.get(ctx, "accounts/me", &out); err != nil {
		return nil, fmt.Errorf("Me: %w", err)
	}
	return &out, nil
}

type CreatePaymentOptions struct {
	// Товары транзакции.
	Items []Product `json:"items"`
	// URL страницы, на которую попадет пользователь после оплаты.
	RedirectURL string `json:"redirectUrl"`
	// URL, куда наш сервер направит запрос, чтобы оповестить ваш сервер об успешной оплате.
	WebhookURL string `json:"webhookUrl"`
	// Сюда можно поместить любые полезные данные. Ограничение - 100 символов.
	Payload string `json:"data"`
}

func NewCreatePaymentOptions(items []Product, redirectURL, webhookURL string) CreatePaymentOptions {
	return CreatePaymentOptions{
		Items:       items,
		RedirectURL: redirectURL,
		WebhookURL:  webhookURL,
	}
}

func (o CreatePaymentOptions) SetPayload(payload string) CreatePaymentOptions {
	o.Payload = payload
	return o
}

func (c *Client) CreatePayment(ctx context.Context, opts CreatePaymentOptions) (*Payment, error) {
	var out Payment
	if err := c.post(ctx, "payments", opts, &out); err != nil {
		return nil, fmt.Errorf("CreatePayment: %w", err)
	}
	return &out, nil
}

type CreateTransactionOptions struct {
	// Номер карты получателя.
	Receiver string `json:"receiver"`
	// Количество АРов для перевода.
	Amount int `json:"amount"`
	// Комментарий для перевода. Ограничение - 64 символа.
	Comment string `json:"comment"`
}

func NewCreateTransactionOptions(receiver string, amount int) CreateTransactionOptions {
	return CreateTransactionOptions{
		Receiver: receiver,
		Amount:   amount,
		Comment:  "",
	}
}

func (o CreateTransactionOptions) SetComment(comment string) CreateTransactionOptions {
	o.Comment = comment
	return o
}

type CreateTransactionResponse struct {
	// Баланс карты после транзакции.
	Balance int `json:"balance"`
}

func (c *Client) CreateTransaction(ctx context.Context, opts CreateTransactionOptions) (*CreateTransactionResponse, error) {
	var out CreateTransactionResponse
	if err := c.post(ctx, "transactions", opts, &out); err != nil {
		return nil, fmt.Errorf("CreateTransaction: %w", err)
	}
	return &out, nil
}

type UpdateWebhookOptions struct {
	URL string `json:"url"`
}

func NewUpdateWebhookOptions(url string) UpdateWebhookOptions {
	return UpdateWebhookOptions{
		URL: url,
	}
}

type UpdateWebhookResponse struct {
	// Уникальный ID карты.
	ID string `json:"id"`
	// Обновленный webhook карты.
	Webhook string `json:"webhook"`
}

// На вебхук будут отправляться все новые транзакции связанные с картой.
// Данные будут отправлены через POST запрос.
func (c *Client) UpdateWebhook(ctx context.Context, opts UpdateWebhookOptions) (*UpdateWebhookResponse, error) {
	var out UpdateWebhookResponse
	if err := c.put(ctx, "card/webhook", opts, &out); err != nil {
		return nil, fmt.Errorf("UpdateWebhook: %w", err)
	}
	return &out, nil
}
