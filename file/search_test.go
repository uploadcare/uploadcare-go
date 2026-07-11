package file

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uploadcare/uploadcare-go/v2/internal/uctest"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

func matchByID(id string) SearchMatch {
	return SearchMatch{ID: id}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	t.Run("method_path_and_body_query_split", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/files/search/", r.URL.Path)
			assert.Equal(t, "50", r.URL.Query().Get("limit"))
			assert.Equal(t, "appdata", r.URL.Query().Get("include"))

			body := uctest.ParseJSONMap(t, uctest.ReadBody(t, r))
			assert.JSONEq(t, `"invoice"`, string(body["query"]))
			for _, k := range []string{"limit", "offset", "include"} {
				_, ok := body[k]
				assert.Falsef(t, ok, "%q must not appear in the body", k)
			}

			uctest.RespondJSON(t, w, searchPage{
				Results: []SearchMatch{{
					ID:               "uuid-1",
					OriginalFileName: "invoice.pdf",
					Size:             100,
				}},
				Total: 1,
			})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			res, err := svc.Search(context.Background(), SearchParams{
				Query:   "invoice",
				Limit:   ucare.Uint64(50),
				Include: ucare.String(SearchIncludeAppData),
			})
			require.NoError(t, err)
			require.True(t, res.Next())

			m, err := res.ReadResult()
			require.NoError(t, err)
			assert.Equal(t, "uuid-1", m.ID)
			assert.Equal(t, uint64(1), res.Total())
		})
	})

	t.Run("body_conditions", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := uctest.ParseJSONMap(t, uctest.ReadBody(t, r))

			assert.JSONEq(t, `{"original_filename":"report"}`, string(body["phrase"]))
			assert.JSONEq(t, `{"uuid":["x"],"metadata[camera]":["Canon"]}`, string(body["exact"]))
			assert.JSONEq(t, `{"gte":"2024-01-02T03:04:05Z"}`, string(body["datetime_uploaded"]))
			assert.JSONEq(t, `{"gte":1000,"lte":5000}`, string(body["size"]))
			assert.JSONEq(t, `{"any":["urgent"],"none":["archived"]}`, string(body["tags"]))
			assert.JSONEq(t, `false`, string(body["is_image"]))
			assert.JSONEq(t, `["datetime_uploaded","-size"]`, string(body["sort"]))

			for _, k := range []string{"query", "fuzziness"} {
				_, ok := body[k]
				assert.Falsef(t, ok, "%q must be omitted", k)
			}

			uctest.RespondJSON(t, w, searchPage{Total: 0})
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			_, err := svc.Search(context.Background(), SearchParams{
				Phrase:           &SearchPhrase{OriginalFilename: "report"},
				Exact:            map[string][]string{SearchExactKeyUUID: {"x"}, "metadata[camera]": {"Canon"}},
				DatetimeUploaded: &SearchDatetime{Gte: ucare.Time(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC))},
				Size:             &SearchSize{Gte: ucare.Uint64(1000), Lte: ucare.Uint64(5000)},
				Tags:             &SearchTags{Any: []string{"urgent"}, None: []string{"archived"}},
				IsImage:          ucare.Bool(false),
				Sort:             []SearchSort{SortByUploadedAt, SortBySizeDesc},
			})
			require.NoError(t, err)
		})
	})

	t.Run("decodes_full_file_info_highlight_and_appdata", func(t *testing.T) {
		t.Parallel()

		const raw = `{
			"total": 1,
			"next": null,
			"results": [{
				"uuid": "uuid-1",
				"original_filename": "invoice.pdf",
				"size": 145212,
				"mime_type": "application/pdf",
				"is_image": false,
				"is_ready": true,
				"datetime_uploaded": "2024-06-01T12:00:00Z",
				"datetime_stored": "2024-06-01T12:00:01Z",
				"datetime_removed": null,
				"original_file_url": "https://ucarecdn.com/uuid-1/invoice.pdf",
				"url": "https://api.uploadcare.com/files/uuid-1/",
				"source": null,
				"variations": null,
				"content_info": {"mime": {"mime": "application/pdf", "type": "application", "subtype": "pdf"}},
				"metadata": {"camera": "Canon"},
				"highlight": {
					"original_filename": ["<em>inv</em>oice.pdf"],
					"metadata": {"camera": "<em>Canon</em>"}
				},
				"appdata": {"uc_clamav_virus_scan": {"data": {"infected": false}}}
			}]
		}`

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(raw))
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			res, err := svc.Search(context.Background(), SearchParams{Query: "invoice"})
			require.NoError(t, err)
			require.True(t, res.Next())

			m, err := res.ReadResult()
			require.NoError(t, err)
			assert.Equal(t, "uuid-1", m.ID)
			assert.Equal(t, "invoice.pdf", m.OriginalFileName)
			assert.Equal(t, uint64(145212), m.Size)
			assert.Equal(t, "application/pdf", m.MimeType)
			assert.False(t, m.IsImage)
			assert.True(t, m.IsReady)
			require.NotNil(t, m.UploadedAt)
			require.NotNil(t, m.StoredAt)
			assert.Nil(t, m.RemovedAt)
			require.NotNil(t, m.OriginalFileURL)
			assert.Equal(t, "https://ucarecdn.com/uuid-1/invoice.pdf", *m.OriginalFileURL)
			assert.Equal(t, "https://api.uploadcare.com/files/uuid-1/", m.URL)
			assert.Equal(t, "Canon", m.Metadata["camera"])
			require.NotNil(t, m.ContentInfo)
			require.NotNil(t, m.ContentInfo.Mime)
			assert.Equal(t, "application/pdf", m.ContentInfo.Mime.Mime)
			assert.Nil(t, m.Variations)

			require.NotNil(t, m.Highlight)
			assert.Equal(t, []string{"<em>inv</em>oice.pdf"}, m.Highlight.OriginalFileName)
			assert.Equal(t, "<em>Canon</em>", m.Highlight.Metadata["camera"])

			require.Contains(t, m.AppData, "uc_clamav_virus_scan")
			assert.JSONEq(t, `{"data":{"infected":false}}`, string(m.AppData["uc_clamav_virus_scan"]))
		})
	})

	t.Run("rewrites_cdn_base", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uctest.RespondJSON(t, w, searchPage{
				Results: []SearchMatch{{
					ID:              rewriteUUID,
					OriginalFileURL: ucare.String(legacyURL),
				}},
				Total: 1,
			})
		}), func(t *testing.T, srv *httptest.Server) {
			c := uctest.NewServerClient(srv)
			c.CDN = rewriteCDN
			svc := NewService(c)
			res, err := svc.Search(context.Background(), SearchParams{Query: "invoice"})
			require.NoError(t, err)
			require.True(t, res.Next())

			m, err := res.ReadResult()
			require.NoError(t, err)
			require.NotNil(t, m.OriginalFileURL)
			assert.Equal(t, expectedRewritten, *m.OriginalFileURL)
		})
	})

	t.Run("paginates", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			switch off := r.URL.Query().Get("offset"); off {
			case "":
				// The next URL advertises offset=50, which does NOT equal the
				// number of results returned; the client must follow the URL
				// rather than derive the offset from the result count.
				uctest.RespondJSON(t, w, map[string]any{
					"total":   100,
					"next":    "https://api.uploadcare.com/files/search/?limit=2&offset=50&include=appdata",
					"results": []SearchMatch{matchByID("uuid-1"), matchByID("uuid-2")},
				})
			case "50":
				// The follow-up request must carry the next URL's query verbatim
				// (limit, include) and re-send the body conditions.
				assert.Equal(t, "2", r.URL.Query().Get("limit"))
				assert.Equal(t, "appdata", r.URL.Query().Get("include"))
				body := uctest.ParseJSONMap(t, uctest.ReadBody(t, r))
				assert.JSONEq(t, `"invoice"`, string(body["query"]))

				uctest.RespondJSON(t, w, map[string]any{
					"total":   100,
					"next":    nil,
					"results": []SearchMatch{matchByID("uuid-3")},
				})
			default:
				t.Fatalf("unexpected offset %q (did not follow next URL verbatim)", off)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			res, err := svc.Search(context.Background(), SearchParams{
				Query:   "invoice",
				Limit:   ucare.Uint64(2),
				Include: ucare.String(SearchIncludeAppData),
			})
			require.NoError(t, err)

			var ids []string
			for res.Next() {
				m, err := res.ReadResult()
				require.NoError(t, err)
				ids = append(ids, m.ID)
			}

			assert.Equal(t, []string{"uuid-1", "uuid-2", "uuid-3"}, ids)
			assert.Equal(t, uint64(100), res.Total())
			assert.False(t, res.Next())
		})
	})

	t.Run("follows_next_through_empty_page", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch off := r.URL.Query().Get("offset"); off {
			case "":
				uctest.RespondJSON(t, w, map[string]any{
					"total":   5,
					"next":    "https://api.uploadcare.com/files/search/?limit=2&offset=2",
					"results": []SearchMatch{matchByID("uuid-1"), matchByID("uuid-2")},
				})
			case "2":
				// Empty page with a next pointer: happens when the matches in
				// this window were stale and got filtered out server-side.
				uctest.RespondJSON(t, w, map[string]any{
					"total":   5,
					"next":    "https://api.uploadcare.com/files/search/?limit=2&offset=4",
					"results": []SearchMatch{},
				})
			case "4":
				uctest.RespondJSON(t, w, map[string]any{
					"total":   5,
					"next":    nil,
					"results": []SearchMatch{matchByID("uuid-3")},
				})
			default:
				t.Fatalf("unexpected offset %q (pagination did not terminate)", off)
			}
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			res, err := svc.Search(context.Background(), SearchParams{
				Query: "invoice",
				Limit: ucare.Uint64(2),
			})
			require.NoError(t, err)

			var ids []string
			for res.Next() {
				m, err := res.ReadResult()
				require.NoError(t, err)
				ids = append(ids, m.ID)
			}

			assert.Equal(t, []string{"uuid-1", "uuid-2", "uuid-3"}, ids)
			assert.False(t, res.Next())
		})
	})

	t.Run("validation_short_circuits_before_request", func(t *testing.T) {
		t.Parallel()

		uctest.WithHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}), func(t *testing.T, srv *httptest.Server) {
			svc := NewService(uctest.NewServerClient(srv))
			_, err := svc.Search(context.Background(), SearchParams{})
			assert.ErrorIs(t, err, ErrSearchNoCondition)
		})
	})
}
