package app

import (
	"context"
	"net/http"
)

func (a *App) seats(ctx context.Context, u User, id int64) ([]Seat, error) {
	f, e := a.accessible(ctx, u, "voo", id)
	if e != nil {
		return nil, e
	}
	craft, e := one(ctx, a.DB, "aeronave", num(f["aeronave_id"]))
	if e != nil {
		return nil, e
	}
	occupied, e := rows(ctx, a.DB, "SELECT assento FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", id)
	if e != nil {
		return nil, e
	}
	taken := map[int64]bool{}
	for _, r := range occupied {
		taken[num(r["assento"])] = true
	}
	out := []Seat{}
	for i := 1; i <= numInt(craft["capacidade"]); i++ {
		out = append(out, Seat{i, taken[int64(i)]})
	}
	return out, nil
}
func (a *App) seatsHandler(w http.ResponseWriter, r *http.Request) {
	seats, e := a.seats(r.Context(), current(r), num(r.URL.Query().Get("voo_id")))
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	a.render(w, "seats", seats)
}
