import type { UserDto } from "@/types/user";
import { api } from "./axiosInstance";

export const usersApi = {
  paged: (params: QueryParams) =>
    api.get<ApiPagedResponse<UserDto>>("/api/users/paged", params),
  get: (id: number) => api.get<ApiResponse<UserDto>>(`/api/users/${id}`),
  delete: (id: number) => api.delete<ApiResponse<void>>(`/api/users/${id}`),
};
