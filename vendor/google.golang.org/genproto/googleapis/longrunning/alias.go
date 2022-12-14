// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by aliasgen. DO NOT EDIT.

// Package longrunning aliases all exported identifiers in package
// "cloud.google.com/go/longrunning/autogen/longrunningpb".
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb.
// Please read https://github.com/googleapis/google-cloud-go/blob/main/migration.md
// for more details.
package longrunning

import (
	src "cloud.google.com/go/longrunning/autogen/longrunningpb"
	grpc "google.golang.org/grpc"
)

// Deprecated: Please use vars in: cloud.google.com/go/longrunning/autogen/longrunningpb
var (
	E_OperationInfo                          = src.E_OperationInfo
	File_google_longrunning_operations_proto = src.File_google_longrunning_operations_proto
)

// The request message for
// [Operations.CancelOperation][google.longrunning.Operations.CancelOperation].
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type CancelOperationRequest = src.CancelOperationRequest

// The request message for
// [Operations.DeleteOperation][google.longrunning.Operations.DeleteOperation].
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type DeleteOperationRequest = src.DeleteOperationRequest

// The request message for
// [Operations.GetOperation][google.longrunning.Operations.GetOperation].
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type GetOperationRequest = src.GetOperationRequest

// The request message for
// [Operations.ListOperations][google.longrunning.Operations.ListOperations].
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type ListOperationsRequest = src.ListOperationsRequest

// The response message for
// [Operations.ListOperations][google.longrunning.Operations.ListOperations].
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type ListOperationsResponse = src.ListOperationsResponse

// This resource represents a long-running operation that is the result of a
// network API call.
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type Operation = src.Operation

// A message representing the message types used by a long-running operation.
// Example: rpc LongRunningRecognize(LongRunningRecognizeRequest) returns
// (google.longrunning.Operation) { option (google.longrunning.operation_info)
// = { response_type: "LongRunningRecognizeResponse" metadata_type:
// "LongRunningRecognizeMetadata" }; }
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type OperationInfo = src.OperationInfo
type Operation_Error = src.Operation_Error
type Operation_Response = src.Operation_Response

// OperationsClient is the client API for Operations service. For semantics
// around ctx use and closing/ending streaming RPCs, please refer to
// https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type OperationsClient = src.OperationsClient

// OperationsServer is the server API for Operations service.
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type OperationsServer = src.OperationsServer

// UnimplementedOperationsServer can be embedded to have forward compatible
// implementations.
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type UnimplementedOperationsServer = src.UnimplementedOperationsServer

// The request message for
// [Operations.WaitOperation][google.longrunning.Operations.WaitOperation].
//
// Deprecated: Please use types in: cloud.google.com/go/longrunning/autogen/longrunningpb
type WaitOperationRequest = src.WaitOperationRequest

// Deprecated: Please use funcs in: cloud.google.com/go/longrunning/autogen/longrunningpb
func NewOperationsClient(cc grpc.ClientConnInterface) OperationsClient {
	return src.NewOperationsClient(cc)
}

// Deprecated: Please use funcs in: cloud.google.com/go/longrunning/autogen/longrunningpb
func RegisterOperationsServer(s *grpc.Server, srv OperationsServer) {
	src.RegisterOperationsServer(s, srv)
}
