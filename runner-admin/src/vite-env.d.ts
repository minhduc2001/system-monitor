/// <reference types="vite/client" />

interface ErrorResponse {
  message?: string;
  errorCode?: string;
}

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

interface SuspenseWrapperProps {
  children: ReactElement;
}

interface AsyncWrapperProps {
  loading: boolean;
  fulfilled: boolean;
  error?: unknown;
  children: React.JSX;
}

interface HelmetProps {
  title: string;
  description: string;
}

interface QueryParams {
  limit?: number;
  page?: number;
  search?: string;
  filter?: string;
  sort?: string[];
}

interface MetaResponse {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

interface ApiResponse<T> {
  message: string;
  data: T;
}

interface ApiPagedResponse<T> {
  message: string;
  data: {
    results: T[];
    meta: MetaResponse;
  };
}
