export interface SchemaStore {
  Examples: { [key: string]: any[] };
  Optional: { [key: string]: boolean };
}

export interface ResponseData {
  Headers: SchemaStore;
  Payload: SchemaStore;
}

export interface EndpointData {
  Method: string;
  URL: string;
  RequestHeaders: SchemaStore;
  RequestPayload: SchemaStore;
  URLParameters: SchemaStore;
  ResponseStatuses: { [key: string]: ResponseData };
}

export interface AnalyzerData {
  [key: string]: EndpointData;
} 