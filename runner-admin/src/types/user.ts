export interface UserRoleDto {
  id: number;
  name: string;
}

export interface UserDto {
  id: number;
  name: string;
  email: string;
  active: boolean;
  roles: UserRoleDto[];
  createdAt: string;
  updatedAt: string;
}
