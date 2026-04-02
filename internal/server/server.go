package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/stockyard-dev/stockyard-archive/internal/store"
)

type Server struct {
	db  *store.DB
	mux *http.ServeMux
}

func New(db *store.DB) *Server {
	s := &Server{db: db, mux: http.NewServeMux()}

	// Collections
	s.mux.HandleFunc("GET /api/collections", s.listCollections)
	s.mux.HandleFunc("POST /api/collections", s.createCollection)
	s.mux.HandleFunc("GET /api/collections/{id}", s.getCollection)
	s.mux.HandleFunc("PUT /api/collections/{id}", s.updateCollection)
	s.mux.HandleFunc("DELETE /api/collections/{id}", s.deleteCollection)

	// Clips
	s.mux.HandleFunc("GET /api/clips", s.listClips)
	s.mux.HandleFunc("POST /api/clips", s.saveClip)
	s.mux.HandleFunc("GET /api/clips/{id}", s.getClip)
	s.mux.HandleFunc("PUT /api/clips/{id}", s.updateClip)
	s.mux.HandleFunc("DELETE /api/clips/{id}", s.deleteClip)
	s.mux.HandleFunc("POST /api/clips/{id}/status", s.setStatus)
	s.mux.HandleFunc("POST /api/clips/{id}/favorite", s.toggleFavorite)

	// Annotations
	s.mux.HandleFunc("GET /api/clips/{id}/annotations", s.listAnnotations)
	s.mux.HandleFunc("POST /api/clips/{id}/annotations", s.createAnnotation)
	s.mux.HandleFunc("DELETE /api/annotations/{id}", s.deleteAnnotation)

	// Meta
	s.mux.HandleFunc("GET /api/tags", s.allTags)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", http.StatusFound)
}

// ── Collections ──

func (s *Server) listCollections(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"collections": orEmpty(s.db.ListCollections())})
}

func (s *Server) createCollection(w http.ResponseWriter, r *http.Request) {
	var c store.Collection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	if c.Name == "" {
		writeErr(w, 400, "name required")
		return
	}
	if err := s.db.CreateCollection(&c); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, s.db.GetCollection(c.ID))
}

func (s *Server) getCollection(w http.ResponseWriter, r *http.Request) {
	c := s.db.GetCollection(r.PathValue("id"))
	if c == nil {
		writeErr(w, 404, "not found")
		return
	}
	writeJSON(w, 200, c)
}

func (s *Server) updateCollection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ex := s.db.GetCollection(id)
	if ex == nil {
		writeErr(w, 404, "not found")
		return
	}
	var c store.Collection
	json.NewDecoder(r.Body).Decode(&c)
	if c.Name == "" {
		c.Name = ex.Name
	}
	if c.Icon == "" {
		c.Icon = ex.Icon
	}
	s.db.UpdateCollection(id, &c)
	writeJSON(w, 200, s.db.GetCollection(id))
}

func (s *Server) deleteCollection(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteCollection(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

// ── Clips ──

func (s *Server) listClips(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	f := store.ClipFilter{
		CollectionID: q.Get("collection_id"),
		Status:       q.Get("status"),
		Tag:          q.Get("tag"),
		Favorite:     q.Get("favorite"),
		Search:       q.Get("search"),
		SortBy:       q.Get("sort"),
		Limit:        limit,
		Offset:       offset,
	}
	if f.Status == "" {
		f.Status = "all"
	}
	clips, total := s.db.ListClips(f)
	writeJSON(w, 200, map[string]any{"clips": orEmpty(clips), "total": total})
}

func (s *Server) saveClip(w http.ResponseWriter, r *http.Request) {
	var c store.Clip
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	if c.Title == "" && c.URL == "" {
		writeErr(w, 400, "title or url required")
		return
	}
	if c.Title == "" {
		c.Title = c.URL
	}
	if err := s.db.SaveClip(&c); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, s.db.GetClip(c.ID))
}

func (s *Server) getClip(w http.ResponseWriter, r *http.Request) {
	c := s.db.GetClip(r.PathValue("id"))
	if c == nil {
		writeErr(w, 404, "not found")
		return
	}
	writeJSON(w, 200, c)
}

func (s *Server) updateClip(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ex := s.db.GetClip(id)
	if ex == nil {
		writeErr(w, 404, "not found")
		return
	}
	var c store.Clip
	json.NewDecoder(r.Body).Decode(&c)
	if c.Title == "" {
		c.Title = ex.Title
	}
	if c.URL == "" {
		c.URL = ex.URL
	}
	if c.Content == "" {
		c.Content = ex.Content
	}
	if c.Tags == nil {
		c.Tags = ex.Tags
	}
	if c.CollectionID == "" {
		c.CollectionID = ex.CollectionID
	}
	s.db.UpdateClip(id, &c)
	writeJSON(w, 200, s.db.GetClip(id))
}

func (s *Server) deleteClip(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteClip(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) setStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		writeErr(w, 400, "status required (inbox, read, archived)")
		return
	}
	s.db.SetStatus(id, req.Status)
	writeJSON(w, 200, s.db.GetClip(id))
}

func (s *Server) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.db.ToggleFavorite(id)
	writeJSON(w, 200, s.db.GetClip(id))
}

// ── Annotations ──

func (s *Server) listAnnotations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"annotations": orEmpty(s.db.ListAnnotations(r.PathValue("id")))})
}

func (s *Server) createAnnotation(w http.ResponseWriter, r *http.Request) {
	clipID := r.PathValue("id")
	if s.db.GetClip(clipID) == nil {
		writeErr(w, 404, "clip not found")
		return
	}
	var a store.Annotation
	json.NewDecoder(r.Body).Decode(&a)
	if a.Note == "" && a.Highlight == "" {
		writeErr(w, 400, "note or highlight required")
		return
	}
	a.ClipID = clipID
	s.db.CreateAnnotation(&a)
	writeJSON(w, 201, a)
}

func (s *Server) deleteAnnotation(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteAnnotation(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

// ── Meta ──

func (s *Server) allTags(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"tags": orEmpty(s.db.AllTags())})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	writeJSON(w, 200, map[string]any{
		"status":  "ok",
		"service": "archive",
		"clips":   st.Total,
		"inbox":   st.Inbox,
	})
}

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
