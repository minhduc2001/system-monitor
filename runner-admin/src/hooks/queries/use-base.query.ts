import { api } from "@/api/axiosInstance";
import { usersApi } from "@/api";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryResult,
} from "@tanstack/react-query";
import type { UserDto } from "@/types/user";

type LocalQueryParams = Record<string, unknown>;

export const useRefetchQueries = () => {
  const queryClient = useQueryClient();
  return async (keys: string[]) => {
    await Promise.all(
      keys.map((key) => queryClient.refetchQueries({ queryKey: [key] }))
    );
  };
};

export const usePaginateModels = <T>(
  path: string,
  params?: LocalQueryParams
) => {
  return useQuery({
    queryKey: [path, "paginate", params],
    queryFn: () => api.get<T>(path, params),
  });
};

export const useUsersPaged = (
  params: QueryParams
): UseQueryResult<ApiPagedResponse<UserDto>> => {
  return useQuery<ApiPagedResponse<UserDto>>({
    queryKey: ["users", "paged", params],
    queryFn: () => usersApi.paged(params),
    placeholderData: keepPreviousData,
  });
};

export const useCreateModel = <ResponseModel, Body>(path: string) => {
  return useMutation({
    mutationFn: async ({ data }: { data: Body }) => {
      return await api.post<ResponseModel>(path, data);
    },
  });
};

export const useUpdateModel = <ResponseModel, Body>(path: string) => {
  return useMutation({
    mutationFn: async ({ id, data }: { id: number | string; data: Body }) => {
      const updateData = { ...data, id };
      return await api.put<ResponseModel>(path + "/" + id, updateData);
    },
  });
};
