package nats

import "github.com/nats-io/nats.go"

func (h *Handler) Register(nc *nats.Conn) error {
	subjects := map[string]nats.MsgHandler{
		"todo.list.create":           h.HandleCreateList,
		"todo.list.get_by_user":      h.HandleGetListsByUser,
		"todo.list.update_title":     h.HandleUpdateListTitle,
		"todo.list.delete":           h.HandleDeleteList,
		"todo.item.create":           h.HandleCreateItem,
		"todo.item.get_all":          h.HandleGetAllItemsFromList,
		"todo.item.toggle_completed": h.HandleToggleItemCompleted,
		"todo.item.delete":           h.HandleDeleteItem,
	}

	for subject, handler := range subjects {
		if _, err := nc.Subscribe(subject, handler); err != nil {
			return err
		}
		h.logger.Info("subscribed to subject", "subject", subject)
	}

	return nil
}
