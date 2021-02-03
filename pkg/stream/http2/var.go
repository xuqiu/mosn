/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http2

import (
	"context"
	"fmt"
	"strings"

	"mosn.io/api"
	mosnctx "mosn.io/pkg/context"
	"mosn.io/pkg/variable"

	"mosn.io/mosn/pkg/protocol"
)

var (
	headerIndex = len(variable.VarPrefixHttp2Header)
	cookieIndex = len(variable.VarPrefixHttp2Cookie)

	builtinVariables = []variable.Variable{
		variable.NewBasicVariable(variable.VarHttp2RequestScheme, nil, schemeGetter, nil, 0),
	}

	prefixVariables = []variable.Variable{
		variable.NewBasicVariable(variable.VarPrefixHttp2Header, nil, headerGetter, nil, 0),
		variable.NewBasicVariable(variable.VarPrefixHttp2Cookie, nil, cookieGetter, nil, 0),
	}
)

func init() {
	// register built-in variables
	for idx := range builtinVariables {
		variable.RegisterVariable(builtinVariables[idx])
	}

	// register prefix variables, like header_xxx/arg_xxx/cookie_xxx
	for idx := range prefixVariables {
		variable.RegisterPrefixVariable(prefixVariables[idx].Name(), prefixVariables[idx])
	}

	// register protocol resource
	variable.RegisterProtocolResource(protocol.HTTP2, api.SCHEME, variable.VarProtocolRequestScheme)
	variable.RegisterProtocolResource(protocol.HTTP2, api.HEADER, variable.VarProtocolRequestHeader)
	variable.RegisterProtocolResource(protocol.HTTP2, api.COOKIE, variable.VarProtocolCookie)
}

func schemeGetter(ctx context.Context, value *variable.IndexedValue, data interface{}) (string, error) {
	scheme, err := variable.GetVariableValue(ctx, variable.VarScheme)
	if err != nil || scheme == "" {
		return variable.ValueNotFound, nil
	}
	return scheme, nil
}

func headerGetter(ctx context.Context, value *variable.IndexedValue, data interface{}) (s string, err error) {
	headers, ok := mosnctx.Get(ctx, mosnctx.ContextKeyDownStreamHeaders).(api.HeaderMap)
	if !ok {
		return variable.ValueNotFound, nil
	}
	headerKey, ok := data.(string)
	if !ok {
		return variable.ValueNotFound, nil
	}

	header, found := headers.Get(headerKey[headerIndex:])
	if !found {
		return variable.ValueNotFound, nil
	}

	return header, nil
}

func cookieGetter(ctx context.Context, value *variable.IndexedValue, data interface{}) (s string, err error) {
	headers, ok := mosnctx.Get(ctx, mosnctx.ContextKeyDownStreamHeaders).(api.HeaderMap)
	if !ok {
		return variable.ValueNotFound, nil
	}
	cookieKey, ok := data.(string)
	if !ok {
		return variable.ValueNotFound, nil
	}

	cookiePrefix := fmt.Sprintf("%s=", cookieKey[cookieIndex:])
	cookieString, found := headers.Get("Cookie")
	if !found {
		return variable.ValueNotFound, nil
	}

	for _, cookieKV := range strings.Split(cookieString, ";") {
		kv := strings.TrimSpace(cookieKV)
		if strings.HasPrefix(kv, cookiePrefix) {
			return strings.TrimSpace(strings.TrimPrefix(kv, cookiePrefix)), nil
		}
	}
	return variable.ValueNotFound, nil
}
