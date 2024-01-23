package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerTypes(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"server-mixed-payload-attrs", testdata.MixedPayloadInBodyDSL, MixedPayloadInBodyServerTypesFile},
		{"server-multiple-methods", testdata.MultipleMethodsDSL, MultipleMethodsServerTypesFile},
		{"server-payload-extend-validate", testdata.PayloadExtendedValidateDSL, PayloadExtendedValidateServerTypesFile},
		{"server-result-type-validate", testdata.ResultTypeValidateDSL, ResultTypeValidateServerTypesFile},
		{"server-with-result-collection", testdata.ResultWithResultCollectionDSL, ResultWithResultCollectionServerTypesFile},
		{"server-with-result-view", testdata.ResultWithResultViewDSL, ResultWithResultViewServerTypesFile},
		{"server-empty-error-response-body", testdata.EmptyErrorResponseBodyDSL, ""},
		{"server-with-error-custom-pkg", testdata.WithErrorCustomPkgDSL, WithErrorCustomPkgServerTypesFile},
		{"server-body-custom-name", testdata.PayloadBodyCustomNameDSL, BodyCustomNameServerTypesFile},
		{"server-path-custom-name", testdata.PayloadPathCustomNameDSL, PathCustomNameServerTypesFile},
		{"server-query-custom-name", testdata.PayloadQueryCustomNameDSL, QueryCustomNameServerTypesFile},
		{"server-header-custom-name", testdata.PayloadHeaderCustomNameDSL, HeaderCustomNameServerTypesFile},
		{"server-cookie-custom-name", testdata.PayloadCookieCustomNameDSL, CookieCustomNameServerTypesFile},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := serverType(genpkg, expr.Root.API.HTTP.Services[0], make(map[string]struct{}))
			var buf bytes.Buffer
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			assert.Equal(t, c.Code, code)
		})
	}
}

const MixedPayloadInBodyServerTypesFile = `// MethodARequestBody is the type of the "ServiceMixedPayloadInBody" service
// "MethodA" endpoint HTTP request body.
type MethodARequestBody struct {
	Any    any                  ` + "`" + `form:"any,omitempty" json:"any,omitempty" xml:"any,omitempty"` + "`" + `
	Array  []float32            ` + "`" + `form:"array,omitempty" json:"array,omitempty" xml:"array,omitempty"` + "`" + `
	Map    map[uint]any         ` + "`" + `form:"map,omitempty" json:"map,omitempty" xml:"map,omitempty"` + "`" + `
	Object *BPayloadRequestBody ` + "`" + `form:"object,omitempty" json:"object,omitempty" xml:"object,omitempty"` + "`" + `
	DupObj *BPayloadRequestBody ` + "`" + `form:"dup_obj,omitempty" json:"dup_obj,omitempty" xml:"dup_obj,omitempty"` + "`" + `
}

// BPayloadRequestBody is used to define fields on request body types.
type BPayloadRequestBody struct {
	Int   *int   ` + "`" + `form:"int,omitempty" json:"int,omitempty" xml:"int,omitempty"` + "`" + `
	Bytes []byte ` + "`" + `form:"bytes,omitempty" json:"bytes,omitempty" xml:"bytes,omitempty"` + "`" + `
}

// NewMethodAAPayload builds a ServiceMixedPayloadInBody service MethodA
// endpoint payload.
func NewMethodAAPayload(body *MethodARequestBody) *servicemixedpayloadinbody.APayload {
	v := &servicemixedpayloadinbody.APayload{
		Any: body.Any,
	}
	v.Array = make([]float32, len(body.Array))
	for i, val := range body.Array {
		v.Array[i] = val
	}
	if body.Map != nil {
		v.Map = make(map[uint]any, len(body.Map))
		for key, val := range body.Map {
			tk := key
			tv := val
			v.Map[tk] = tv
		}
	}
	v.Object = unmarshalBPayloadRequestBodyToServicemixedpayloadinbodyBPayload(body.Object)
	if body.DupObj != nil {
		v.DupObj = unmarshalBPayloadRequestBodyToServicemixedpayloadinbodyBPayload(body.DupObj)
	}

	return v
}

// ValidateMethodARequestBody runs the validations defined on MethodARequestBody
func ValidateMethodARequestBody(body *MethodARequestBody) (err error) {
	if body.Array == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("array", "body"))
	}
	if body.Object == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("object", "body"))
	}
	if body.Object != nil {
		if err2 := ValidateBPayloadRequestBody(body.Object); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if body.DupObj != nil {
		if err2 := ValidateBPayloadRequestBody(body.DupObj); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateBPayloadRequestBody runs the validations defined on
// BPayloadRequestBody
func ValidateBPayloadRequestBody(body *BPayloadRequestBody) (err error) {
	if body.Int == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("int", "body"))
	}
	return
}
`

const MultipleMethodsServerTypesFile = `// MethodARequestBody is the type of the "ServiceMultipleMethods" service
// "MethodA" endpoint HTTP request body.
type MethodARequestBody struct {
	A *string ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
}

// MethodBRequestBody is the type of the "ServiceMultipleMethods" service
// "MethodB" endpoint HTTP request body.
type MethodBRequestBody struct {
	A *string              ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
	B *string              ` + "`" + `form:"b,omitempty" json:"b,omitempty" xml:"b,omitempty"` + "`" + `
	C *APayloadRequestBody ` + "`" + `form:"c,omitempty" json:"c,omitempty" xml:"c,omitempty"` + "`" + `
}

// APayloadRequestBody is used to define fields on request body types.
type APayloadRequestBody struct {
	A *string ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
}

// NewMethodAAPayload builds a ServiceMultipleMethods service MethodA endpoint
// payload.
func NewMethodAAPayload(body *MethodARequestBody) *servicemultiplemethods.APayload {
	v := &servicemultiplemethods.APayload{
		A: body.A,
	}

	return v
}

// NewMethodBPayloadType builds a ServiceMultipleMethods service MethodB
// endpoint payload.
func NewMethodBPayloadType(body *MethodBRequestBody) *servicemultiplemethods.PayloadType {
	v := &servicemultiplemethods.PayloadType{
		A: *body.A,
		B: body.B,
	}
	v.C = unmarshalAPayloadRequestBodyToServicemultiplemethodsAPayload(body.C)

	return v
}

// ValidateMethodARequestBody runs the validations defined on MethodARequestBody
func ValidateMethodARequestBody(body *MethodARequestBody) (err error) {
	if body.A != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.a", *body.A, "patterna"))
	}
	return
}

// ValidateMethodBRequestBody runs the validations defined on MethodBRequestBody
func ValidateMethodBRequestBody(body *MethodBRequestBody) (err error) {
	if body.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "body"))
	}
	if body.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "body"))
	}
	if body.A != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.a", *body.A, "patterna"))
	}
	if body.B != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.b", *body.B, "patternb"))
	}
	if body.C != nil {
		if err2 := ValidateAPayloadRequestBody(body.C); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAPayloadRequestBody runs the validations defined on
// APayloadRequestBody
func ValidateAPayloadRequestBody(body *APayloadRequestBody) (err error) {
	if body.A != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.a", *body.A, "patterna"))
	}
	return
}
`

const PayloadExtendedValidateServerTypesFile = `// MethodQueryStringExtendedValidatePayloadRequestBody is the type of the
// "ServiceQueryStringExtendedValidatePayload" service
// "MethodQueryStringExtendedValidatePayload" endpoint HTTP request body.
type MethodQueryStringExtendedValidatePayloadRequestBody struct {
	Body *string ` + "`" + `form:"body,omitempty" json:"body,omitempty" xml:"body,omitempty"` + "`" + `
}

// NewMethodQueryStringExtendedValidatePayloadPayload builds a
// ServiceQueryStringExtendedValidatePayload service
// MethodQueryStringExtendedValidatePayload endpoint payload.
func NewMethodQueryStringExtendedValidatePayloadPayload(body *MethodQueryStringExtendedValidatePayloadRequestBody, q string, h int) *servicequerystringextendedvalidatepayload.MethodQueryStringExtendedValidatePayloadPayload {
	v := &servicequerystringextendedvalidatepayload.MethodQueryStringExtendedValidatePayloadPayload{
		Body: *body.Body,
	}
	v.Q = q
	v.H = h

	return v
}

// ValidateMethodQueryStringExtendedValidatePayloadRequestBody runs the
// validations defined on MethodQueryStringExtendedValidatePayloadRequestBody
func ValidateMethodQueryStringExtendedValidatePayloadRequestBody(body *MethodQueryStringExtendedValidatePayloadRequestBody) (err error) {
	if body.Body == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("body", "body"))
	}
	return
}
`

const ResultTypeValidateServerTypesFile = `// MethodResultTypeValidateResponseBody is the type of the
// "ServiceResultTypeValidate" service "MethodResultTypeValidate" endpoint HTTP
// response body.
type MethodResultTypeValidateResponseBody struct {
	A *string ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
}

// NewMethodResultTypeValidateResponseBody builds the HTTP response body from
// the result of the "MethodResultTypeValidate" endpoint of the
// "ServiceResultTypeValidate" service.
func NewMethodResultTypeValidateResponseBody(res *serviceresulttypevalidate.ResultType) *MethodResultTypeValidateResponseBody {
	body := &MethodResultTypeValidateResponseBody{
		A: res.A,
	}
	return body
}
`

const ResultWithResultCollectionServerTypesFile = `// MethodResultWithResultCollectionResponseBody is the type of the
// "ServiceResultWithResultCollection" service
// "MethodResultWithResultCollection" endpoint HTTP response body.
type MethodResultWithResultCollectionResponseBody struct {
	A *ResulttypeResponseBody ` + "`" + `form:"a,omitempty" json:"a,omitempty" xml:"a,omitempty"` + "`" + `
}

// ResulttypeResponseBody is used to define fields on response body types.
type ResulttypeResponseBody struct {
	X RtCollectionResponseBody ` + "`" + `form:"x,omitempty" json:"x,omitempty" xml:"x,omitempty"` + "`" + `
}

// RtCollectionResponseBody is used to define fields on response body types.
type RtCollectionResponseBody []*RtResponseBody

// RtResponseBody is used to define fields on response body types.
type RtResponseBody struct {
	X *string ` + "`" + `form:"x,omitempty" json:"x,omitempty" xml:"x,omitempty"` + "`" + `
}

// NewMethodResultWithResultCollectionResponseBody builds the HTTP response
// body from the result of the "MethodResultWithResultCollection" endpoint of
// the "ServiceResultWithResultCollection" service.
func NewMethodResultWithResultCollectionResponseBody(res *serviceresultwithresultcollection.MethodResultWithResultCollectionResult) *MethodResultWithResultCollectionResponseBody {
	body := &MethodResultWithResultCollectionResponseBody{}
	if res.A != nil {
		body.A = marshalServiceresultwithresultcollectionResulttypeToResulttypeResponseBody(res.A)
	}
	return body
}
`

const ResultWithResultViewServerTypesFile = `// MethodResultWithResultViewResponseBodyFull is the type of the
// "ServiceResultWithResultView" service "MethodResultWithResultView" endpoint
// HTTP response body.
type MethodResultWithResultViewResponseBodyFull struct {
	Name *string         ` + "`" + `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"` + "`" + `
	Rt   *RtResponseBody ` + "`" + `form:"rt,omitempty" json:"rt,omitempty" xml:"rt,omitempty"` + "`" + `
}

// RtResponseBody is used to define fields on response body types.
type RtResponseBody struct {
	X *string ` + "`" + `form:"x,omitempty" json:"x,omitempty" xml:"x,omitempty"` + "`" + `
}

// NewMethodResultWithResultViewResponseBodyFull builds the HTTP response body
// from the result of the "MethodResultWithResultView" endpoint of the
// "ServiceResultWithResultView" service.
func NewMethodResultWithResultViewResponseBodyFull(res *serviceresultwithresultviewviews.ResulttypeView) *MethodResultWithResultViewResponseBodyFull {
	body := &MethodResultWithResultViewResponseBodyFull{
		Name: res.Name,
	}
	if res.Rt != nil {
		body.Rt = marshalServiceresultwithresultviewviewsRtViewToRtResponseBody(res.Rt)
	}
	return body
}
`

const WithErrorCustomPkgServerTypesFile = `// MethodWithErrorCustomPkgErrorNameResponseBody is the type of the
// "ServiceWithErrorCustomPkg" service "MethodWithErrorCustomPkg" endpoint HTTP
// response body for the "error_name" error.
type MethodWithErrorCustomPkgErrorNameResponseBody struct {
	Name string ` + "`" + `form:"name" json:"name" xml:"name"` + "`" + `
}

// NewMethodWithErrorCustomPkgErrorNameResponseBody builds the HTTP response
// body from the result of the "MethodWithErrorCustomPkg" endpoint of the
// "ServiceWithErrorCustomPkg" service.
func NewMethodWithErrorCustomPkgErrorNameResponseBody(res *custom.CustomError) *MethodWithErrorCustomPkgErrorNameResponseBody {
	body := &MethodWithErrorCustomPkgErrorNameResponseBody{
		Name: res.Name,
	}
	return body
}
`

const BodyCustomNameServerTypesFile = `// MethodBodyCustomNameRequestBody is the type of the "ServiceBodyCustomName"
// service "MethodBodyCustomName" endpoint HTTP request body.
type MethodBodyCustomNameRequestBody struct {
	Body *string ` + "`" + `form:"b,omitempty" json:"b,omitempty" xml:"b,omitempty"` + "`" + `
}

// NewMethodBodyCustomNamePayload builds a ServiceBodyCustomName service
// MethodBodyCustomName endpoint payload.
func NewMethodBodyCustomNamePayload(body *MethodBodyCustomNameRequestBody) *servicebodycustomname.MethodBodyCustomNamePayload {
	v := &servicebodycustomname.MethodBodyCustomNamePayload{
		Body: body.Body,
	}

	return v
}
`
const PathCustomNameServerTypesFile = `// NewMethodPathCustomNamePayload builds a ServicePathCustomName service
// MethodPathCustomName endpoint payload.
func NewMethodPathCustomNamePayload(p string) *servicepathcustomname.MethodPathCustomNamePayload {
	v := &servicepathcustomname.MethodPathCustomNamePayload{}
	v.Path = p

	return v
}
`
const QueryCustomNameServerTypesFile = `// NewMethodQueryCustomNamePayload builds a ServiceQueryCustomName service
// MethodQueryCustomName endpoint payload.
func NewMethodQueryCustomNamePayload(q *string) *servicequerycustomname.MethodQueryCustomNamePayload {
	v := &servicequerycustomname.MethodQueryCustomNamePayload{}
	v.Query = q

	return v
}
`

const HeaderCustomNameServerTypesFile = `// NewMethodHeaderCustomNamePayload builds a ServiceHeaderCustomName service
// MethodHeaderCustomName endpoint payload.
func NewMethodHeaderCustomNamePayload(h *string) *serviceheadercustomname.MethodHeaderCustomNamePayload {
	v := &serviceheadercustomname.MethodHeaderCustomNamePayload{}
	v.Header = h

	return v
}
`

const CookieCustomNameServerTypesFile = `// NewMethodCookieCustomNamePayload builds a ServiceCookieCustomName service
// MethodCookieCustomName endpoint payload.
func NewMethodCookieCustomNamePayload(c2 *string) *servicecookiecustomname.MethodCookieCustomNamePayload {
	v := &servicecookiecustomname.MethodCookieCustomNamePayload{}
	v.Cookie = c2

	return v
}
`
