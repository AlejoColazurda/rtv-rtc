export interface CompanyInfo {
  name: string;
  company: string;
  cuit: string;
  address: string;
  email: string;
  logo?: string;
}

export interface Item {
  qty: number;
  description: string;
  unitPrice: number;
  total: number;
}

export interface LayoutConfig {
  pageSize: 'A4' | 'Letter' | 'A5' | 'custom';
  widthMm: number;
  heightMm: number;
  marginTop: number;
  marginBottom: number;
  marginLeft: number;
  marginRight: number;
  colQtyWidth: number;
  colDescWidth: number;
  colPriceWidth: number;
  colTotalWidth: number;
  fontSizeScale: number;
}

export interface Invitation {
  id?: string;
  type: 'remito' | 'orden';
  docNumber: string;
  emisor: CompanyInfo;
  receptor: CompanyInfo;
  eventDate: string;
  items: Item[];
  comments: string;
  theme: 'classic' | 'coffee' | 'blue' | 'minimal';
  status?: 'pending' | 'accepted' | 'declined';
  signature?: string;
  rejectionReason?: string;
  layout: LayoutConfig;
  createdAt?: string;
}
