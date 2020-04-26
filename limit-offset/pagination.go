package pagination

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	// ParamPageLimit is the value for a page number parameter on http request
	ParamPageLimit = "page[limit]"
	// ParamPageOffset is the value for a page size parameter on http request
	ParamPageOffset = "page[offset]"
	// ParamSortBy is the value for the sorting query
	ParamSortBy = "sort"
)

// Paginate will build a new paginated response with the given values
func Paginate(data []interface{}, baseURL string, params Params) Response {
	return Response{
		Data:  buildData(data, params),
		Links: buildLinks(baseURL, params, len(data)),
	}
}

// Response type encapsulates the information related with a paginated response
type Response struct {
	Data  []interface{} `json:"data,omitempty"`
	Links Links         `json:"links"`
}

// Links type encapsulates the information about how we can move through the
// different pages on a paginated reponse
type Links struct {
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
}

// Sort type encapsulates the information needed for order and sort a query, the
// field will have the name column to be sorted and the order will have the value
// of asc or desc
type Sort struct {
	Field string
	Order string
}

// Params type encapsulates the information gathered from the http request
type Params struct {
	Limit  uint
	Offset uint
	Sort   []Sort
}

// SortURL will convert the sort slice into a URL parameters
func (p Params) SortURL() (sortParams string) {
	if len(p.Sort) > 0 {
		sortParams = fmt.Sprintf("%s=", ParamSortBy)
		tmp := []string{}
		for _, s := range p.Sort {
			tmp = append(tmp, fmt.Sprintf("%s.%s", s.Field, s.Order))
		}
		sortParams += strings.Join(tmp, ",")
	}
	return sortParams
}

// Query method will build the part of the SQL query that should be attached to
// the end of the parent query
func (p Params) Query() string {
	// This p.Limit + 1 is the approach for know about the last page without having
	// the extra count query
	query := fmt.Sprintf(" LIMIT %d OFFSET %d ", p.Limit+1, p.Offset)
	if len(p.Sort) > 0 {
		query += "ORDER BY "
		tmp := []string{}
		for _, s := range p.Sort {
			tmp = append(tmp, fmt.Sprintf("%s %s", s.Field, s.Order))
		}
		query += strings.Join(tmp, ",")
	}
	return query
}

// FindParams will find for the pagination params on the request otherwise will
// answer back with the given defaults
func FindParams(req *http.Request, defaultOffset, defaultLimit uint) (Params, error) {
	params := Params{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}
	limit := req.URL.Query().Get(ParamPageLimit)
	offset := req.URL.Query().Get(ParamPageOffset)
	sort := req.URL.Query().Get(ParamSortBy)

	if limit != "" {
		convertedLimit, err := strconv.ParseUint(limit, 10, 32)
		if err != nil {
			return params, err
		}
		params.Limit = uint(convertedLimit)
	}

	if offset != "" {
		convertedOffset, err := strconv.ParseUint(offset, 10, 32)
		if err != nil {
			return params, err
		}
		params.Offset = uint(convertedOffset)
	}

	if sort != "" {
		sortFields := strings.Split(sort, ",")
		for _, field := range sortFields {
			// The format of sort and order values shoulde be something
			// like this name.asc or name.desc
			v := strings.Split(field, ".")
			if len(v) == 2 {
				params.Sort = append(params.Sort, Sort{
					Field: v[0],
					Order: v[1],
				})
			}
		}
	}

	return params, nil
}

// buildLinks function will build the links for navigate through the pages
// using the given criteria
func buildLinks(baseURL string, params Params, dataSize int) (links Links) {
	sortURL := params.SortURL()
	links.First = fmt.Sprintf("%s?%s=%d&%s=%d", baseURL, ParamPageLimit, params.Limit, ParamPageOffset, 0)
	if sortURL != "" {
		links.First += fmt.Sprintf("&%s", sortURL)
	}

	if uint(dataSize) > params.Limit {
		links.Next = fmt.Sprintf("%s?%s=%d&%s=%d", baseURL, ParamPageLimit, params.Limit, ParamPageOffset, params.Offset+params.Limit)
		if sortURL != "" {
			links.Next += fmt.Sprintf("&%s", sortURL)
		}
	}
	if params.Offset > 0 {
		links.Prev = fmt.Sprintf("%s?%s=%d&%s=%d", baseURL, ParamPageLimit, params.Limit, ParamPageOffset, params.Offset-params.Limit)
		if sortURL != "" {
			links.Prev += fmt.Sprintf("&%s", sortURL)
		}
	}
	return links
}

// buildData function will handle the situation of deal with an extra limit for
// avoid extra count query, so in case we should remove the last item we will
// remove it
func buildData(data []interface{}, params Params) []interface{} {
	if uint(len(data)) > params.Limit {
		data = data[:len(data)-1]
	}
	return data
}
