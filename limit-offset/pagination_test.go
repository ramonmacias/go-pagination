package pagination_test

import (
	"net/http"
	"testing"

	pagination "github.com/ramonmacias/go-pagination/limit-offset"
	"github.com/stretchr/testify/assert"
)

func TestFindOffSetAndLimitParams(t *testing.T) {
	expectedDefaultOffset := uint(2)
	expectedDefaultLimit := uint(4)

	expectedOffset := uint(10)
	expectedLimit := uint(5)

	req, err := http.NewRequest(
		http.MethodGet,
		"app.quicka.co/api/sample?page[limit]=5&page[offset]=10",
		nil,
	)
	assert.Nil(t, err)

	params, err := pagination.FindParams(req, expectedDefaultOffset, expectedDefaultLimit)
	assert.Nil(t, err)
	assert.Equal(t, expectedLimit, params.Limit)
	assert.Equal(t, expectedOffset, params.Offset)
	assert.Equal(t, 0, len(params.Sort))

	defaultReq, err := http.NewRequest(
		http.MethodGet,
		"app.quicka.co/api/sample",
		nil,
	)
	assert.Nil(t, err)

	defaultValueParams, err := pagination.FindParams(defaultReq, expectedDefaultOffset, expectedDefaultLimit)
	assert.Nil(t, err)
	assert.Equal(t, expectedDefaultLimit, defaultValueParams.Limit)
	assert.Equal(t, expectedDefaultOffset, defaultValueParams.Offset)
	assert.Equal(t, 0, len(defaultValueParams.Sort))
}

func TestFindSortAndOrderParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want []pagination.Sort
	}{
		{
			name: "Should return an empty slice",
			url:  "app.quicka.co/api/sample",
			want: nil,
		},
		{
			name: "Should return an empty slice due wrong format",
			url:  "app.quicka.co/api/simple?sort=asc(name)",
			want: nil,
		},
		{
			name: "Should return a slice with one item",
			url:  "app.quicka.co/api/simple?sort=name.asc",
			want: []pagination.Sort{
				{
					Field: "name",
					Order: "asc",
				},
			},
		},
		{
			name: "Should return a slice with two items",
			url:  "app.quicka.co/api/simple?sort=name.asc,second_name.desc",
			want: []pagination.Sort{
				{
					Field: "name",
					Order: "asc",
				},
				{
					Field: "second_name",
					Order: "desc",
				},
			},
		},
		{
			name: "Should avoid mallformed sort value",
			url:  "app.quicka.co/api/simple?sort=name.asc,second_name.desc,asc(muz)",
			want: []pagination.Sort{
				{
					Field: "name",
					Order: "asc",
				},
				{
					Field: "second_name",
					Order: "desc",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodGet,
				tt.url,
				nil,
			)
			assert.Nil(t, err)
			params, err := pagination.FindParams(req, 0, 0)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, params.Sort)
		})
	}
}

func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name string
		args pagination.Params
		want string
	}{
		{
			name: "Default case",
			args: pagination.Params{},
			want: " LIMIT 1 OFFSET 0 ",
		},
		{
			name: "Specific value case",
			args: pagination.Params{
				Limit:  uint(10),
				Offset: uint(20),
			},
			want: " LIMIT 11 OFFSET 20 ",
		},
		{
			name: "Params with one order case",
			args: pagination.Params{
				Limit:  uint(10),
				Offset: uint(20),
				Sort: []pagination.Sort{
					{
						Field: "first_name",
						Order: "asc",
					},
				},
			},
			want: " LIMIT 11 OFFSET 20 ORDER BY first_name asc",
		},
		{
			name: "Params with two order cases",
			args: pagination.Params{
				Limit:  uint(2),
				Offset: uint(34),
				Sort: []pagination.Sort{
					{
						Field: "last_name",
						Order: "asc",
					},
					{
						Field: "created_at",
						Order: "desc",
					},
				},
			},
			want: " LIMIT 3 OFFSET 34 ORDER BY last_name asc,created_at desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.args.Query())
		})
	}
}

func TestSortURLMethod(t *testing.T) {
	tests := []struct {
		name string
		args pagination.Params
		want string
	}{
		{
			name: "Empty sort params",
			args: pagination.Params{},
			want: "",
		},
		{
			name: "Sort size 1",
			args: pagination.Params{
				Sort: []pagination.Sort{
					{
						Field: "first_name",
						Order: "asc",
					},
				},
			},
			want: "sort=first_name.asc",
		},
		{
			name: "Sort size 2",
			args: pagination.Params{
				Sort: []pagination.Sort{
					{
						Field: "first_name",
						Order: "asc",
					},
					{
						Field: "created_at",
						Order: "desc",
					},
				},
			},
			want: "sort=first_name.asc,created_at.desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.args.SortURL())
		})
	}
}

func TestPaginatedResponseBuilder(t *testing.T) {
	type testArgs struct {
		data    []string
		baseURL string
		params  pagination.Params
	}
	tests := []struct {
		name string
		args testArgs
		want pagination.Links
	}{
		{
			name: "First page",
			args: testArgs{
				data:    []string{"sample", "sample2", "sample3", "sample4", "sample5", "sample6"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  5,
					Offset: 0,
				},
			},
			want: pagination.Links{
				First: "/sample?page[limit]=5&page[offset]=0",
				Next:  "/sample?page[limit]=5&page[offset]=5",
			},
		},
		{
			name: "Intermediate page",
			args: testArgs{
				data:    []string{"sample", "sample2", "sample3", "sample4", "sample5", "sample6"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  5,
					Offset: 10,
				},
			},
			want: pagination.Links{
				First: "/sample?page[limit]=5&page[offset]=0",
				Next:  "/sample?page[limit]=5&page[offset]=15",
				Prev:  "/sample?page[limit]=5&page[offset]=5",
			},
		},
		{
			name: "Last page",
			args: testArgs{
				data:    []string{"sample"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  5,
					Offset: 10,
				},
			},
			want: pagination.Links{
				First: "/sample?page[limit]=5&page[offset]=0",
				Prev:  "/sample?page[limit]=5&page[offset]=5",
			},
		},
		{
			name: "Intermediate page with sort params",
			args: testArgs{
				data:    []string{"sample", "sample2", "sample3", "sample4", "sample5", "sample6"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  5,
					Offset: 10,
					Sort: []pagination.Sort{
						{
							Field: "first_name",
							Order: "asc",
						},
						{
							Field: "created_at",
							Order: "desc",
						},
					},
				},
			},
			want: pagination.Links{
				First: "/sample?page[limit]=5&page[offset]=0&sort=first_name.asc,created_at.desc",
				Next:  "/sample?page[limit]=5&page[offset]=15&sort=first_name.asc,created_at.desc",
				Prev:  "/sample?page[limit]=5&page[offset]=5&sort=first_name.asc,created_at.desc",
			},
		},
		{
			name: "First page without next one",
			args: testArgs{
				data:    []string{"sample", "sample2"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  5,
					Offset: 0,
				},
			},
			want: pagination.Links{
				First: "/sample?page[limit]=5&page[offset]=0",
			},
		},
		{
			name: "Last page with full items",
			args: testArgs{
				data:    []string{"sample", "sample2", "sample3", "sample4", "sample5"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  5,
					Offset: 5,
				},
			},
			want: pagination.Links{
				First: "/sample?page[limit]=5&page[offset]=0",
				Prev:  "/sample?page[limit]=5&page[offset]=0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make([]interface{}, len(tt.args.data))
			for i, v := range tt.args.data {
				s[i] = v
			}
			assert.Equal(t, tt.want, pagination.Paginate(s, tt.args.baseURL, tt.args.params).Links)
		})
	}
}

func TestPaginatedDataSize(t *testing.T) {
	type testArgs struct {
		data    []string
		baseURL string
		params  pagination.Params
	}
	tests := []struct {
		name string
		args testArgs
		want int
	}{
		{
			name: "Standard full size page",
			args: testArgs{
				data:    []string{"sample", "sample3", "sample4"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  2,
					Offset: 0,
				},
			},
			want: 2,
		},
		{
			name: "Not full size page",
			args: testArgs{
				data:    []string{"sample", "sample3"},
				baseURL: "/sample",
				params: pagination.Params{
					Limit:  2,
					Offset: 0,
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make([]interface{}, len(tt.args.data))
			for i, v := range tt.args.data {
				s[i] = v
			}
			assert.Equal(t, tt.want, len(pagination.Paginate(s, tt.args.baseURL, tt.args.params).Data))
		})
	}
}
