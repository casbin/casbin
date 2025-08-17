// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import "encoding/json"

func (e *Enforcer) GetImplicitPermissionsForUserWithDomain(user string, domain string) ([]byte, error) {
    policy, err := e.GetImplicitPermissionsForUser(user, domain)
    if err != nil {
        return nil, err
    }

    permission := make(map[string][]string)
    for _, p := range policy {
       
        if len(p) < 3 {
            return nil, fmt.Errorf("invalid policy entry: insufficient elements %v", p)
        }
        permission[p[2]] = append(permission[p[2]], p[1])
    }

    b, jsonErr := json.Marshal(permission)
    if jsonErr != nil {
        return nil, jsonErr
    }

    return b, nil
}

