package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	_	= protoimpl.EnforceVersion(20 - protoimpl.MinVersion)

	_	= protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Order struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	ID		string			`protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	CustomerID	string			`protobuf:"bytes,2,opt,name=customerID,proto3" json:"customerID,omitempty"`
	Status		string			`protobuf:"bytes,3,opt,name=Status,proto3" json:"Status,omitempty"`
	Items		[]*Item			`protobuf:"bytes,4,rep,name=Items,proto3" json:"Items,omitempty"`
	PaymentLink	string			`protobuf:"bytes,5,opt,name=PaymentLink,proto3" json:"PaymentLink,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *Order) Reset() {
	*x = Order{}
	mi := &file_api_oms_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Order) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Order) ProtoMessage()	{}

func (x *Order) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Order) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{0}
}

func (x *Order) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

func (x *Order) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

func (x *Order) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Order) GetItems() []*Item {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *Order) GetPaymentLink() string {
	if x != nil {
		return x.PaymentLink
	}
	return ""
}

type GetOrderRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	OrderID		string			`protobuf:"bytes,1,opt,name=OrderID,proto3" json:"OrderID,omitempty"`
	CustomerID	string			`protobuf:"bytes,2,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetOrderRequest) Reset() {
	*x = GetOrderRequest{}
	mi := &file_api_oms_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetOrderRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetOrderRequest) ProtoMessage()	{}

func (x *GetOrderRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetOrderRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{1}
}

func (x *GetOrderRequest) GetOrderID() string {
	if x != nil {
		return x.OrderID
	}
	return ""
}

func (x *GetOrderRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

type GetCustomerOrdersRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetCustomerOrdersRequest) Reset() {
	*x = GetCustomerOrdersRequest{}
	mi := &file_api_oms_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCustomerOrdersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCustomerOrdersRequest) ProtoMessage()	{}

func (x *GetCustomerOrdersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetCustomerOrdersRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{2}
}

func (x *GetCustomerOrdersRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

type GetCustomerOrdersResponse struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	Orders		[]*Order		`protobuf:"bytes,1,rep,name=Orders,proto3" json:"Orders,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetCustomerOrdersResponse) Reset() {
	*x = GetCustomerOrdersResponse{}
	mi := &file_api_oms_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCustomerOrdersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCustomerOrdersResponse) ProtoMessage()	{}

func (x *GetCustomerOrdersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetCustomerOrdersResponse) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{3}
}

func (x *GetCustomerOrdersResponse) GetOrders() []*Order {
	if x != nil {
		return x.Orders
	}
	return nil
}

type Item struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	ID		string			`protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Name		string			`protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	Quantity	int32			`protobuf:"varint,3,opt,name=Quantity,proto3" json:"Quantity,omitempty"`
	PriceID		string			`protobuf:"bytes,4,opt,name=PriceID,proto3" json:"PriceID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *Item) Reset() {
	*x = Item{}
	mi := &file_api_oms_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Item) ProtoMessage()	{}

func (x *Item) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Item) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{4}
}

func (x *Item) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

func (x *Item) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Item) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *Item) GetPriceID() string {
	if x != nil {
		return x.PriceID
	}
	return ""
}

type ItemsWithQuantity struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	ID		string			`protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Quantity	int32			`protobuf:"varint,2,opt,name=Quantity,proto3" json:"Quantity,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *ItemsWithQuantity) Reset() {
	*x = ItemsWithQuantity{}
	mi := &file_api_oms_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ItemsWithQuantity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ItemsWithQuantity) ProtoMessage()	{}

func (x *ItemsWithQuantity) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ItemsWithQuantity) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{5}
}

func (x *ItemsWithQuantity) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

func (x *ItemsWithQuantity) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

type CreateOrderRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=customerID,proto3" json:"customerID,omitempty"`
	Items		[]*ItemsWithQuantity	`protobuf:"bytes,2,rep,name=Items,proto3" json:"Items,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *CreateOrderRequest) Reset() {
	*x = CreateOrderRequest{}
	mi := &file_api_oms_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateOrderRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateOrderRequest) ProtoMessage()	{}

func (x *CreateOrderRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CreateOrderRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{6}
}

func (x *CreateOrderRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

func (x *CreateOrderRequest) GetItems() []*ItemsWithQuantity {
	if x != nil {
		return x.Items
	}
	return nil
}

type CheckIfItemIsInStockRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	Items		[]*ItemsWithQuantity	`protobuf:"bytes,1,rep,name=Items,proto3" json:"Items,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *CheckIfItemIsInStockRequest) Reset() {
	*x = CheckIfItemIsInStockRequest{}
	mi := &file_api_oms_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CheckIfItemIsInStockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckIfItemIsInStockRequest) ProtoMessage()	{}

func (x *CheckIfItemIsInStockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CheckIfItemIsInStockRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{7}
}

func (x *CheckIfItemIsInStockRequest) GetItems() []*ItemsWithQuantity {
	if x != nil {
		return x.Items
	}
	return nil
}

type CheckIfItemIsInStockResponse struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	InStock		bool			`protobuf:"varint,1,opt,name=InStock,proto3" json:"InStock,omitempty"`
	Items		[]*Item			`protobuf:"bytes,2,rep,name=Items,proto3" json:"Items,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *CheckIfItemIsInStockResponse) Reset() {
	*x = CheckIfItemIsInStockResponse{}
	mi := &file_api_oms_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CheckIfItemIsInStockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckIfItemIsInStockResponse) ProtoMessage()	{}

func (x *CheckIfItemIsInStockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CheckIfItemIsInStockResponse) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{8}
}

func (x *CheckIfItemIsInStockResponse) GetInStock() bool {
	if x != nil {
		return x.InStock
	}
	return false
}

func (x *CheckIfItemIsInStockResponse) GetItems() []*Item {
	if x != nil {
		return x.Items
	}
	return nil
}

type GetItemsRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	ItemIDs		[]string		`protobuf:"bytes,1,rep,name=ItemIDs,proto3" json:"ItemIDs,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetItemsRequest) Reset() {
	*x = GetItemsRequest{}
	mi := &file_api_oms_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemsRequest) ProtoMessage()	{}

func (x *GetItemsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetItemsRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{9}
}

func (x *GetItemsRequest) GetItemIDs() []string {
	if x != nil {
		return x.ItemIDs
	}
	return nil
}

type GetItemsResponse struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	Items		[]*Item			`protobuf:"bytes,1,rep,name=Items,proto3" json:"Items,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetItemsResponse) Reset() {
	*x = GetItemsResponse{}
	mi := &file_api_oms_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemsResponse) ProtoMessage()	{}

func (x *GetItemsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetItemsResponse) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{10}
}

func (x *GetItemsResponse) GetItems() []*Item {
	if x != nil {
		return x.Items
	}
	return nil
}

type RegisterRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	Email		string			`protobuf:"bytes,1,opt,name=Email,proto3" json:"Email,omitempty"`
	Password	string			`protobuf:"bytes,2,opt,name=Password,proto3" json:"Password,omitempty"`
	Name		string			`protobuf:"bytes,3,opt,name=Name,proto3" json:"Name,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	mi := &file_api_oms_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage()	{}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{11}
}

func (x *RegisterRequest) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *RegisterRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *RegisterRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type LoginRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	Email		string			`protobuf:"bytes,1,opt,name=Email,proto3" json:"Email,omitempty"`
	Password	string			`protobuf:"bytes,2,opt,name=Password,proto3" json:"Password,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *LoginRequest) Reset() {
	*x = LoginRequest{}
	mi := &file_api_oms_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LoginRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LoginRequest) ProtoMessage()	{}

func (x *LoginRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*LoginRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{12}
}

func (x *LoginRequest) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *LoginRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type AuthResponse struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	Token		string			`protobuf:"bytes,1,opt,name=Token,proto3" json:"Token,omitempty"`
	CustomerID	string			`protobuf:"bytes,2,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *AuthResponse) Reset() {
	*x = AuthResponse{}
	mi := &file_api_oms_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthResponse) ProtoMessage()	{}

func (x *AuthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*AuthResponse) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{13}
}

func (x *AuthResponse) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *AuthResponse) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

type GetUserRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetUserRequest) Reset() {
	*x = GetUserRequest{}
	mi := &file_api_oms_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetUserRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserRequest) ProtoMessage()	{}

func (x *GetUserRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetUserRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{14}
}

func (x *GetUserRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

type User struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	ID		string			`protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Email		string			`protobuf:"bytes,2,opt,name=Email,proto3" json:"Email,omitempty"`
	Name		string			`protobuf:"bytes,3,opt,name=Name,proto3" json:"Name,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *User) Reset() {
	*x = User{}
	mi := &file_api_oms_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage()	{}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*User) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{15}
}

func (x *User) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

func (x *User) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *User) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type CartItem struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	ProductID	string			`protobuf:"bytes,1,opt,name=ProductID,proto3" json:"ProductID,omitempty"`
	Quantity	int32			`protobuf:"varint,2,opt,name=Quantity,proto3" json:"Quantity,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *CartItem) Reset() {
	*x = CartItem{}
	mi := &file_api_oms_proto_msgTypes[16]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CartItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CartItem) ProtoMessage()	{}

func (x *CartItem) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[16]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CartItem) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{16}
}

func (x *CartItem) GetProductID() string {
	if x != nil {
		return x.ProductID
	}
	return ""
}

func (x *CartItem) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

type Cart struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	Items		[]*CartItem		`protobuf:"bytes,2,rep,name=Items,proto3" json:"Items,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *Cart) Reset() {
	*x = Cart{}
	mi := &file_api_oms_proto_msgTypes[17]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Cart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cart) ProtoMessage()	{}

func (x *Cart) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[17]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Cart) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{17}
}

func (x *Cart) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

func (x *Cart) GetItems() []*CartItem {
	if x != nil {
		return x.Items
	}
	return nil
}

type AddItemRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	Item		*CartItem		`protobuf:"bytes,2,opt,name=Item,proto3" json:"Item,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *AddItemRequest) Reset() {
	*x = AddItemRequest{}
	mi := &file_api_oms_proto_msgTypes[18]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddItemRequest) ProtoMessage()	{}

func (x *AddItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[18]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*AddItemRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{18}
}

func (x *AddItemRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

func (x *AddItemRequest) GetItem() *CartItem {
	if x != nil {
		return x.Item
	}
	return nil
}

type RemoveItemRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	ProductID	string			`protobuf:"bytes,2,opt,name=ProductID,proto3" json:"ProductID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *RemoveItemRequest) Reset() {
	*x = RemoveItemRequest{}
	mi := &file_api_oms_proto_msgTypes[19]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveItemRequest) ProtoMessage()	{}

func (x *RemoveItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[19]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RemoveItemRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{19}
}

func (x *RemoveItemRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

func (x *RemoveItemRequest) GetProductID() string {
	if x != nil {
		return x.ProductID
	}
	return ""
}

type GetCartRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *GetCartRequest) Reset() {
	*x = GetCartRequest{}
	mi := &file_api_oms_proto_msgTypes[20]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCartRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCartRequest) ProtoMessage()	{}

func (x *GetCartRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[20]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetCartRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{20}
}

func (x *GetCartRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

type ClearCartRequest struct {
	state		protoimpl.MessageState	`protogen:"open.v1"`
	CustomerID	string			`protobuf:"bytes,1,opt,name=CustomerID,proto3" json:"CustomerID,omitempty"`
	unknownFields	protoimpl.UnknownFields
	sizeCache	protoimpl.SizeCache
}

func (x *ClearCartRequest) Reset() {
	*x = ClearCartRequest{}
	mi := &file_api_oms_proto_msgTypes[21]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClearCartRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClearCartRequest) ProtoMessage()	{}

func (x *ClearCartRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_oms_proto_msgTypes[21]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ClearCartRequest) Descriptor() ([]byte, []int) {
	return file_api_oms_proto_rawDescGZIP(), []int{21}
}

func (x *ClearCartRequest) GetCustomerID() string {
	if x != nil {
		return x.CustomerID
	}
	return ""
}

var File_api_oms_proto protoreflect.FileDescriptor

const file_api_oms_proto_rawDesc = "" +
	"\n" +
	"\rapi/oms.proto\x12\x03api\"\x92\x01\n" +
	"\x05Order\x12\x0e\n" +
	"\x02ID\x18\x01 \x01(\tR\x02ID\x12\x1e\n" +
	"\n" +
	"customerID\x18\x02 \x01(\tR\n" +
	"customerID\x12\x16\n" +
	"\x06Status\x18\x03 \x01(\tR\x06Status\x12\x1f\n" +
	"\x05Items\x18\x04 \x03(\v2\t.api.ItemR\x05Items\x12 \n" +
	"\vPaymentLink\x18\x05 \x01(\tR\vPaymentLink\"K\n" +
	"\x0fGetOrderRequest\x12\x18\n" +
	"\aOrderID\x18\x01 \x01(\tR\aOrderID\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x02 \x01(\tR\n" +
	"CustomerID\":\n" +
	"\x18GetCustomerOrdersRequest\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID\"?\n" +
	"\x19GetCustomerOrdersResponse\x12\"\n" +
	"\x06Orders\x18\x01 \x03(\v2\n" +
	".api.OrderR\x06Orders\"`\n" +
	"\x04Item\x12\x0e\n" +
	"\x02ID\x18\x01 \x01(\tR\x02ID\x12\x12\n" +
	"\x04Name\x18\x02 \x01(\tR\x04Name\x12\x1a\n" +
	"\bQuantity\x18\x03 \x01(\x05R\bQuantity\x12\x18\n" +
	"\aPriceID\x18\x04 \x01(\tR\aPriceID\"?\n" +
	"\x11ItemsWithQuantity\x12\x0e\n" +
	"\x02ID\x18\x01 \x01(\tR\x02ID\x12\x1a\n" +
	"\bQuantity\x18\x02 \x01(\x05R\bQuantity\"b\n" +
	"\x12CreateOrderRequest\x12\x1e\n" +
	"\n" +
	"customerID\x18\x01 \x01(\tR\n" +
	"customerID\x12,\n" +
	"\x05Items\x18\x02 \x03(\v2\x16.api.ItemsWithQuantityR\x05Items\"K\n" +
	"\x1bCheckIfItemIsInStockRequest\x12,\n" +
	"\x05Items\x18\x01 \x03(\v2\x16.api.ItemsWithQuantityR\x05Items\"Y\n" +
	"\x1cCheckIfItemIsInStockResponse\x12\x18\n" +
	"\aInStock\x18\x01 \x01(\bR\aInStock\x12\x1f\n" +
	"\x05Items\x18\x02 \x03(\v2\t.api.ItemR\x05Items\"+\n" +
	"\x0fGetItemsRequest\x12\x18\n" +
	"\aItemIDs\x18\x01 \x03(\tR\aItemIDs\"3\n" +
	"\x10GetItemsResponse\x12\x1f\n" +
	"\x05Items\x18\x01 \x03(\v2\t.api.ItemR\x05Items\"W\n" +
	"\x0fRegisterRequest\x12\x14\n" +
	"\x05Email\x18\x01 \x01(\tR\x05Email\x12\x1a\n" +
	"\bPassword\x18\x02 \x01(\tR\bPassword\x12\x12\n" +
	"\x04Name\x18\x03 \x01(\tR\x04Name\"@\n" +
	"\fLoginRequest\x12\x14\n" +
	"\x05Email\x18\x01 \x01(\tR\x05Email\x12\x1a\n" +
	"\bPassword\x18\x02 \x01(\tR\bPassword\"D\n" +
	"\fAuthResponse\x12\x14\n" +
	"\x05Token\x18\x01 \x01(\tR\x05Token\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x02 \x01(\tR\n" +
	"CustomerID\"0\n" +
	"\x0eGetUserRequest\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID\"@\n" +
	"\x04User\x12\x0e\n" +
	"\x02ID\x18\x01 \x01(\tR\x02ID\x12\x14\n" +
	"\x05Email\x18\x02 \x01(\tR\x05Email\x12\x12\n" +
	"\x04Name\x18\x03 \x01(\tR\x04Name\"D\n" +
	"\bCartItem\x12\x1c\n" +
	"\tProductID\x18\x01 \x01(\tR\tProductID\x12\x1a\n" +
	"\bQuantity\x18\x02 \x01(\x05R\bQuantity\"K\n" +
	"\x04Cart\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID\x12#\n" +
	"\x05Items\x18\x02 \x03(\v2\r.api.CartItemR\x05Items\"S\n" +
	"\x0eAddItemRequest\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID\x12!\n" +
	"\x04Item\x18\x02 \x01(\v2\r.api.CartItemR\x04Item\"Q\n" +
	"\x11RemoveItemRequest\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID\x12\x1c\n" +
	"\tProductID\x18\x02 \x01(\tR\tProductID\"0\n" +
	"\x0eGetCartRequest\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID\"2\n" +
	"\x10ClearCartRequest\x12\x1e\n" +
	"\n" +
	"CustomerID\x18\x01 \x01(\tR\n" +
	"CustomerID2\xeb\x01\n" +
	"\fOrderService\x122\n" +
	"\vCreateOrder\x12\x17.api.CreateOrderRequest\x1a\n" +
	".api.Order\x12,\n" +
	"\bGetOrder\x12\x14.api.GetOrderRequest\x1a\n" +
	".api.Order\x12R\n" +
	"\x11GetCustomerOrders\x12\x1d.api.GetCustomerOrdersRequest\x1a\x1e.api.GetCustomerOrdersResponse\x12%\n" +
	"\vUpdateOrder\x12\n" +
	".api.Order\x1a\n" +
	".api.Order2\xa4\x01\n" +
	"\fStockService\x12[\n" +
	"\x14CheckIfItemIsInStock\x12 .api.CheckIfItemIsInStockRequest\x1a!.api.CheckIfItemIsInStockResponse\x127\n" +
	"\bGetItems\x12\x14.api.GetItemsRequest\x1a\x15.api.GetItemsResponse2\x9c\x01\n" +
	"\vUserService\x123\n" +
	"\bRegister\x12\x14.api.RegisterRequest\x1a\x11.api.AuthResponse\x12-\n" +
	"\x05Login\x12\x11.api.LoginRequest\x1a\x11.api.AuthResponse\x12)\n" +
	"\aGetUser\x12\x13.api.GetUserRequest\x1a\t.api.User2\xc3\x01\n" +
	"\vCartService\x12)\n" +
	"\aAddItem\x12\x13.api.AddItemRequest\x1a\t.api.Cart\x12/\n" +
	"\n" +
	"RemoveItem\x12\x16.api.RemoveItemRequest\x1a\t.api.Cart\x12)\n" +
	"\aGetCart\x12\x13.api.GetCartRequest\x1a\t.api.Cart\x12-\n" +
	"\tClearCart\x12\x15.api.ClearCartRequest\x1a\t.api.CartB$Z\"github.com/jaydeepjr28/commons/apib\x06proto3"

var (
	file_api_oms_proto_rawDescOnce	sync.Once
	file_api_oms_proto_rawDescData	[]byte
)

func file_api_oms_proto_rawDescGZIP() []byte {
	file_api_oms_proto_rawDescOnce.Do(func() {
		file_api_oms_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_oms_proto_rawDesc), len(file_api_oms_proto_rawDesc)))
	})
	return file_api_oms_proto_rawDescData
}

var file_api_oms_proto_msgTypes = make([]protoimpl.MessageInfo, 22)
var file_api_oms_proto_goTypes = []any{
	(*Order)(nil),
	(*GetOrderRequest)(nil),
	(*GetCustomerOrdersRequest)(nil),
	(*GetCustomerOrdersResponse)(nil),
	(*Item)(nil),
	(*ItemsWithQuantity)(nil),
	(*CreateOrderRequest)(nil),
	(*CheckIfItemIsInStockRequest)(nil),
	(*CheckIfItemIsInStockResponse)(nil),
	(*GetItemsRequest)(nil),
	(*GetItemsResponse)(nil),
	(*RegisterRequest)(nil),
	(*LoginRequest)(nil),
	(*AuthResponse)(nil),
	(*GetUserRequest)(nil),
	(*User)(nil),
	(*CartItem)(nil),
	(*Cart)(nil),
	(*AddItemRequest)(nil),
	(*RemoveItemRequest)(nil),
	(*GetCartRequest)(nil),
	(*ClearCartRequest)(nil),
}
var file_api_oms_proto_depIdxs = []int32{
	4,
	0,
	5,
	5,
	4,
	4,
	16,
	16,
	6,
	1,
	2,
	0,
	7,
	9,
	11,
	12,
	14,
	18,
	19,
	20,
	21,
	0,
	0,
	3,
	0,
	8,
	10,
	13,
	13,
	15,
	17,
	17,
	17,
	17,
	21,
	8,
	8,
	8,
	0,
}

func init()	{ file_api_oms_proto_init() }
func file_api_oms_proto_init() {
	if File_api_oms_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath:	reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor:	unsafe.Slice(unsafe.StringData(file_api_oms_proto_rawDesc), len(file_api_oms_proto_rawDesc)),
			NumEnums:	0,
			NumMessages:	22,
			NumExtensions:	0,
			NumServices:	4,
		},
		GoTypes:		file_api_oms_proto_goTypes,
		DependencyIndexes:	file_api_oms_proto_depIdxs,
		MessageInfos:		file_api_oms_proto_msgTypes,
	}.Build()
	File_api_oms_proto = out.File
	file_api_oms_proto_goTypes = nil
	file_api_oms_proto_depIdxs = nil
}
