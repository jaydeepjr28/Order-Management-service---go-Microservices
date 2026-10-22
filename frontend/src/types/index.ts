export type UserRole = 'customer' | 'admin';

export interface DecodedToken {
  customer_id: string;
  role: UserRole;
  token_type: 'access' | 'refresh';
  exp: number;
  iat: number;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  customerID: string;
}

export interface UserSession {
  customerID: string;
  role: UserRole;
  accessToken: string;
}

export interface StockItem {
  ID: string;
  Name: string;
  Quantity: number;
  PriceID?: string;
  Price?: number; // Price in paise (e.g. 19900 = ₹199.00)
}

export interface CartItem {
  ProductID: string;
  Quantity: number;
}

export interface CartResponse {
  CustomerID: string;
  Items: CartItem[] | null;
}

export interface OrderItem {
  ID: string;
  Name: string;
  Quantity: number;
  PriceID?: string;
  Price?: number; // Price in paise
}

export type OrderStatus = 'pending' | 'paid' | 'ready' | string;

export interface Order {
  ID: string;
  customerID: string;
  Status: OrderStatus;
  Items: OrderItem[];
  PaymentLink?: string;
}

export interface CreateOrderResponse {
  order: Order;
  redirectToURL?: string;
}

export interface OrderStatusResponse {
  status: OrderStatus;
}

export interface AdminAnalyticsResponse {
  message?: string;
  orderVolume?: number;
  totalRevenue?: number;
  statusBreakdown?: Record<string, number>;
}
