package post

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/misgorod/co-dev/auth"
	"github.com/misgorod/co-dev/common"
	"go.mongodb.org/mongo-driver/mongo"

	"gopkg.in/go-playground/validator.v9"
)

type Handler struct {
	Client   *mongo.Client
	Validate *validator.Validate
}

type pageOptions struct {
	Offset int `json:"offset" validate:"gte=0"`
	Limit  int `json:"limit" validate:"gte=0"`
}

func (p *Handler) Post(w http.ResponseWriter, r *http.Request) {
	var post *Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		common.RespondError(w, http.StatusBadRequest, "Failed to decode request")
		return
	}
	if err := p.Validate.Struct(post); err != nil {
		switch err.(type) {
		case validator.ValidationErrors:
			common.RespondError(w, http.StatusBadRequest, "Invalid json")
			break
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal validator: %s", err.Error()))
		}
		return
	}
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		common.RespondError(w, http.StatusUnauthorized, "Token is invalid")
	}
	fmt.Println(userID)
	post, err := CreatePost(r.Context(), p.Client, userID, post)
	if err != nil {
		switch err {
		case primitive.ErrInvalidHex:
			common.RespondError(w, http.StatusUnauthorized, "Token is invalid")
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		}
		return
	}

	common.RespondJSON(w, http.StatusCreated, post)
}

func (p *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	var pageOptions *pageOptions
	if err := json.NewDecoder(r.Body).Decode(&pageOptions); err != nil {
		common.RespondError(w, http.StatusBadRequest, "Failed to decode request")
		return
	}
	if err := p.Validate.Struct(pageOptions); err != nil {
		switch err.(type) {
		case validator.ValidationErrors:
			common.RespondError(w, http.StatusBadRequest, "Invalid json")
			break
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		}
		return
	}
	if pageOptions.Limit == 0 || pageOptions.Limit > 50 {
		pageOptions.Limit = 50
	}

	posts, err := GetPosts(r.Context(), p.Client, pageOptions.Offset, pageOptions.Limit)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		return
	}

	common.RespondJSON(w, http.StatusOK, posts)
}

func (p *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		common.RespondError(w, http.StatusBadRequest, "ID not specified")
	}

	post, err := GetPost(r.Context(), p.Client, id)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			common.RespondError(w, http.StatusNotFound, err.Error())
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		}
		return
	}

	common.RespondJSON(w, http.StatusOK, post)
}

func (p *Handler) MemberPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		common.RespondError(w, http.StatusUnauthorized, "Token is invalid")
	}
	postID := chi.URLParam(r, "id")
	if postID == "" {
		common.RespondError(w, http.StatusBadRequest, ErrPostNotFound.Error())
	}
	err := AddMember(r.Context(), p.Client, postID, userID)
	if err != nil {
		switch err {
		case auth.ErrWrongToken:
			common.RespondError(w, http.StatusUnauthorized, err.Error())
		case ErrPostNotFound:
			common.RespondError(w, http.StatusNotFound, err.Error())
		case ErrMemberAlreadyExists:
			common.RespondError(w, http.StatusBadRequest, err.Error())
		case ErrMemberIsAuthor:
			common.RespondError(w, http.StatusConflict, err.Error())
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		}
	}
}

func (p *Handler) MemberDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		common.RespondError(w, http.StatusUnauthorized, "Token is invalid")
	}
	postID := chi.URLParam(r, "id")
	if postID == "" {
		common.RespondError(w, http.StatusBadRequest, ErrPostNotFound.Error())
	}
	err := DeleteMember(r.Context(), p.Client, postID, userID)
	if err != nil {
		switch err {
		case auth.ErrWrongToken:
			common.RespondError(w, http.StatusUnauthorized, err.Error())
		case ErrPostNotFound:
			fallthrough
		case ErrMebmerNotExists:
			common.RespondError(w, http.StatusNotFound, err.Error())
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		}
	}
}

func (p *Handler) PostImage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		common.RespondError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}
	postID := chi.URLParam(r, "id")
	if postID == "" {
		common.RespondError(w, http.StatusBadRequest, ErrPostNotFound.Error())
		return
	}
	post, err := GetPost(r.Context(), p.Client, postID)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			common.RespondError(w, http.StatusNotFound, err.Error())
		default:
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal: %s", err.Error()))
		}
		return
	}
	if post.Author.ID.Hex() != userID {
		common.RespondError(w, http.StatusForbidden, ErrNotAnAuthor.Error())
		return
	}
	r.ParseMultipartForm(16 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		common.RespondError(w, http.StatusBadRequest, ErrUploadFile.Error())
	}
	defer file.Close()
	//TODO
}
