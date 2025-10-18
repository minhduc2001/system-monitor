import { Button, Card, Form, Input, Typography, message } from "antd";
import { useNavigate } from "react-router-dom";
import useAuthStore, { type User } from "@stores/useAuthStore";

export default function Login() {
  const navigate = useNavigate();
  const { setToken, setUser } = useAuthStore();

  const onFinish = async (values: { username: string; password: string }) => {
    try {
      const fakeToken = "fake-token";
      const fakeUser: User = {
        id: "1",
        username: values.username,
        email: `${values.username}@example.com`,
        fullName: values.username,
        role: "admin",
        permissions: ["users:view"],
      };
      setToken(fakeToken);
      setUser(fakeUser);
      message.success("Đăng nhập thành công");
      navigate("/", { replace: true });
    } catch {
      message.error("Đăng nhập thất bại");
    }
  };

  return (
    <div style={{ minHeight: "100vh", display: "grid", placeItems: "center" }}>
      <Card style={{ width: 360 }}>
        <Typography.Title level={3} style={{ textAlign: "center" }}>
          Đăng nhập
        </Typography.Title>
        <Form layout="vertical" onFinish={onFinish}>
          <Form.Item name="username" label="Tên đăng nhập" rules={[{ required: true }]}>
            <Input placeholder="admin" />
          </Form.Item>
          <Form.Item name="password" label="Mật khẩu" rules={[{ required: true }]}>
            <Input.Password placeholder="••••••" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block>
            Đăng nhập
          </Button>
        </Form>
      </Card>
    </div>
  );
}
