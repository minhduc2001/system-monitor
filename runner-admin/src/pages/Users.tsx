import { useMemo, useState, type ChangeEvent } from "react";
import { Button, Card, Flex, Input, Select, Space, Table, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useUsersPaged } from "@/hooks/queries/use-base.query";
import type { UserDto, UserRoleDto } from "@/types/user";

export default function Users() {
  const [query, setQuery] = useState<QueryParams & { isActive?: string }>({
    search: "",
    isActive: undefined,
    page: 1,
    limit: 10,
  });

  const { data: usersResponse, isLoading } = useUsersPaged(query);

  const columns: ColumnsType<UserDto> = [
    { title: "ID", dataIndex: "id", key: "id", width: 80 },
    { title: "Name", dataIndex: "name", key: "name" },
    { title: "Email", dataIndex: "email", key: "email" },
    {
      title: "Active",
      dataIndex: "active",
      key: "active",
      width: 120,
      render: (v: boolean) =>
        v ? <Tag color="green">Active</Tag> : <Tag color="red">Inactive</Tag>,
    },
    {
      title: "Roles",
      key: "roles",
      render: (_: unknown, r: UserDto) => (
        <Space size={4} wrap>
          {r.roles?.map((role) => (
            <Tag key={role.id}>{role.name}</Tag>
          ))}
        </Space>
      ),
    },
    { title: "Created", dataIndex: "createdAt", key: "createdAt", width: 200 },
  ];

  const results: UserDto[] = usersResponse?.data?.results ?? [];
  const customerResults = useMemo<UserDto[]>(
    () =>
      results.filter((u: UserDto) =>
        u.roles?.some((r: UserRoleDto) => r.name === "USER")
      ),
    [results]
  );
  const meta = usersResponse?.data?.meta;

  return (
    <Card title="Users">
      <Flex gap={8} align="center" wrap style={{ marginBottom: 12 }}>
        <Input
          placeholder="Search name/email"
          value={query.search}
          onChange={(e: ChangeEvent<HTMLInputElement>) =>
            setQuery((q) => ({ ...q, search: e.target.value }))
          }
          allowClear
          style={{ width: 260 }}
        />
        <Select
          placeholder="Active"
          value={query.isActive}
          onChange={(v: string | undefined) =>
            setQuery((q) => ({ ...q, isActive: v }))
          }
          allowClear
          style={{ width: 160 }}
          options={[
            { label: "All", value: undefined },
            { label: "Active", value: "true" },
            { label: "Inactive", value: "false" },
          ]}
        />
        <Button onClick={() => setQuery((q) => ({ ...q, page: 1 }))}>
          Apply
        </Button>
      </Flex>

      <Table<UserDto>
        rowKey="id"
        loading={isLoading}
        columns={columns}
        dataSource={customerResults}
        pagination={{
          current: (meta?.page ?? 0) + 1,
          pageSize: meta?.limit ?? query.limit,
          total: meta?.total ?? 0,
          onChange: (p: number, ps: number) => {
            setQuery((q) => ({ ...q, page: p, limit: ps }));
          },
          showSizeChanger: true,
        }}
      />
    </Card>
  );
}
