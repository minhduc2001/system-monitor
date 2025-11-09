import { useEffect, useState } from "react";
import {
  Form,
  Input,
  Select,
  Button,
  Card,
  Row,
  Col,
  InputNumber,
  Switch,
  message,
  Space,
} from "antd";
import { useNavigate, useParams } from "react-router-dom";
import {
  useProject,
  useCreateProject,
  useUpdateProject,
} from "@/hooks/queries/use-project.query";
import { useProjectGroups } from "@/hooks/queries/use-project.query";
import type {
  CreateProjectRequest,
  UpdateProjectRequest,
} from "@/types/project";
import PathPicker, {
  type DetectedService,
} from "@/components/projects/PathPicker";
import ImportMultipleServices from "@/components/projects/ImportMultipleServices";

const { TextArea } = Input;
const { Option } = Select;

interface ProjectFormProps {
  mode: "create" | "edit";
}

export default function ProjectForm({ mode }: ProjectFormProps) {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [form] = Form.useForm();

  // Watch path field to keep PathPicker in sync
  const pathValue = Form.useWatch("path", form);

  const { data: project, isLoading: isLoadingProject } = useProject(
    mode === "edit" && id ? Number(id) : 0
  );
  const { data: groups = [] } = useProjectGroups();
  const createProject = useCreateProject();
  const updateProject = useUpdateProject();

  const [detectedService, setDetectedService] = useState<{
    name: string;
    type: string;
    path: string;
    command: string;
  } | null>(null);
  const [detectedServices, setDetectedServices] = useState<DetectedService[]>(
    []
  );
  const [showImportServicesModal, setShowImportServicesModal] = useState(false);

  useEffect(() => {
    if (mode === "edit" && project) {
      form.setFieldsValue(project);
    }
  }, [mode, project, form]);

  const handleSubmit = async (
    values: CreateProjectRequest | UpdateProjectRequest
  ) => {
    try {
      if (mode === "create") {
        await createProject.mutateAsync(values as CreateProjectRequest);
        message.success("Project created successfully");
        navigate("/projects");
      } else if (id) {
        await updateProject.mutateAsync({
          id: Number(id),
          data: values as UpdateProjectRequest,
        });
        message.success("Project updated successfully");
        navigate(`/projects/${id}`);
      }
    } catch (error: any) {
      message.error(error?.response?.data?.error || "Failed to save project");
    }
  };

  const handleServiceDetected = (service: DetectedService) => {
    setDetectedService(service);
    // Auto-fill form fields based on detected service
    const currentValues = form.getFieldsValue();
    form.setFieldsValue({
      ...currentValues,
      path: service.path,
      command: service.command || currentValues.command,
      type: service.type || currentValues.type,
      name: service.name || currentValues.name,
    });
  };

  if (mode === "edit" && isLoadingProject) {
    return <div>Loading...</div>;
  }

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      initialValues={{
        type: "backend",
        environment: "development",
        auto_restart: false,
        max_restarts: 3,
      }}
    >
      <Row gutter={16}>
        <Col span={24}>
          <Card title="Basic Information" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="name"
                  label="Project Name"
                  rules={[
                    { required: true, message: "Please enter project name" },
                  ]}
                >
                  <Input placeholder="e.g., User Service" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="type"
                  label="Service Type"
                  rules={[{ required: true }]}
                >
                  <Select>
                    <Option value="backend">Backend</Option>
                    <Option value="frontend">Frontend</Option>
                    <Option value="worker">Worker</Option>
                    <Option value="database">Database</Option>
                    <Option value="queue">Queue</Option>
                    <Option value="other">Other</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="description" label="Description">
              <TextArea rows={3} placeholder="Project description..." />
            </Form.Item>
            <Form.Item name="group_id" label="Project Group">
              <Select allowClear placeholder="Select a group">
                {groups.map((group) => (
                  <Option key={group.id} value={group.id}>
                    {group.name}
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Card>
        </Col>

        <Col span={24}>
          <Card title="Path & Execution" style={{ marginBottom: 16 }}>
            <Form.Item
              name="path"
              label="Project Path"
              rules={[
                { required: true, message: "Please enter project path" },
                {
                  validator: (_, value) => {
                    if (!value) return Promise.resolve();

                    // Check for placeholder
                    if (value.includes("YourUsername")) {
                      return Promise.reject(
                        new Error(
                          'Please replace "YourUsername" with your actual username (e.g., /Users/ducnm/Documents/...). Run "whoami" in Terminal to find your username.'
                        )
                      );
                    }

                    // Basic validation - check if it looks like an absolute path
                    // Windows path: C:\path or \\server\path
                    const isWindowsPath = /^([a-zA-Z]:\\|\\\\).+/.test(value);
                    // Unix path: /path/to/project
                    const isUnixPath = value.startsWith("/");

                    if (!isWindowsPath && !isUnixPath) {
                      return Promise.reject(
                        new Error(
                          "Please enter an absolute path (e.g., /path/to/project or C:\\path\\to\\project)"
                        )
                      );
                    }
                    return Promise.resolve();
                  },
                },
              ]}
              extra={
                <div>
                  <div
                    style={{ marginTop: 4, fontSize: "12px", color: "#888" }}
                  >
                    💡 <strong>Workflow:</strong>
                    <ol style={{ margin: "4px 0", paddingLeft: "20px" }}>
                      <li>
                        Enter the parent folder path (e.g.,{" "}
                        <code>/Users/username/Documents/da-ban-quan-ao</code>)
                      </li>
                      <li>
                        Click "Select Folder" and choose a subfolder (e.g.,{" "}
                        <code>frontend</code> or <code>backend</code>)
                      </li>
                      <li>
                        The path will be automatically combined (e.g.,{" "}
                        <code>
                          /Users/username/Documents/da-ban-quan-ao/frontend
                        </code>
                        )
                      </li>
                      <li>Click "Detect" to auto-configure services</li>
                    </ol>
                  </div>
                </div>
              }
            >
              <PathPicker
                value={pathValue || ""}
                onChange={(value) => {
                  console.log("PathPicker onChange called with:", value);
                  form.setFieldsValue({ path: value });
                }}
                onServiceDetected={handleServiceDetected}
                onMultipleServicesDetected={(services) => {
                  setDetectedServices(services);
                  setShowImportServicesModal(true);
                }}
                placeholder={
                  window.navigator.platform.includes("Win")
                    ? "C:\\Users\\username\\projects\\my-project"
                    : "/Users/username/projects/my-project"
                }
              />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="command"
                  label="Start Command"
                  extra={
                    detectedService?.command
                      ? `Auto-detected: ${detectedService.command}`
                      : "Leave empty to auto-detect"
                  }
                >
                  <Input placeholder="e.g., npm start, go run main.go" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="working_dir" label="Working Directory">
                  <Input placeholder="Leave empty to use project path" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="args" label="Arguments">
              <Input placeholder="Additional command arguments" />
            </Form.Item>
          </Card>
        </Col>

        <Col span={24}>
          <Card title="Network & Ports" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="port" label="Port">
                  <InputNumber
                    min={1}
                    max={65535}
                    style={{ width: "100%" }}
                    placeholder="e.g., 3000"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="ports" label="Multiple Ports (JSON)">
                  <Input placeholder='e.g., ["3000", "3001"]' />
                </Form.Item>
              </Col>
            </Row>
          </Card>
        </Col>

        <Col span={24}>
          <Card title="Environment" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="environment" label="Environment">
                  <Select>
                    <Option value="development">Development</Option>
                    <Option value="staging">Staging</Option>
                    <Option value="production">Production</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="env_file" label="Environment File">
                  <Input placeholder=".env" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="env_vars" label="Environment Variables (JSON)">
              <TextArea rows={4} placeholder='{"KEY": "value"}' />
            </Form.Item>
          </Card>
        </Col>

        <Col span={24}>
          <Card title="Health Check" style={{ marginBottom: 16 }}>
            <Form.Item name="health_check_url" label="Health Check URL">
              <Input placeholder="http://localhost:3000/health" />
            </Form.Item>
          </Card>
        </Col>

        <Col span={24}>
          <Card title="Advanced Settings" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="auto_restart"
                  valuePropName="checked"
                  label="Auto Restart"
                >
                  <Switch />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="max_restarts" label="Max Restarts">
                  <InputNumber min={1} max={10} style={{ width: "100%" }} />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="cpu_limit" label="CPU Limit">
                  <Input placeholder="e.g., 500m" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="memory_limit" label="Memory Limit">
                  <Input placeholder="e.g., 512Mi" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="editor" label="Editor">
                  <Input placeholder="e.g., vscode, intellij" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="editor_args" label="Editor Arguments">
                  <Input placeholder="Editor arguments" />
                </Form.Item>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      <ImportMultipleServices
        services={detectedServices}
        visible={showImportServicesModal}
        onClose={() => {
          setShowImportServicesModal(false);
          setDetectedServices([]);
        }}
        onComplete={() => {
          navigate("/projects");
        }}
      />

      <Space>
        <Button
          type="primary"
          htmlType="submit"
          loading={createProject.isPending || updateProject.isPending}
        >
          {mode === "create" ? "Create" : "Update"}
        </Button>
        <Button onClick={() => navigate("/projects")}>Cancel</Button>
      </Space>
    </Form>
  );
}
