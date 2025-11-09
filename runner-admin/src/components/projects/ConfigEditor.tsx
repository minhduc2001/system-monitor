import { useState, useEffect } from 'react';
import { Card, Button, Space, Select, message, Spin, Alert } from 'antd';
import { SaveOutlined, ReloadOutlined, DownloadOutlined } from '@ant-design/icons';
import AceEditor from 'react-ace';
import 'ace-builds/src-noconflict/mode-yaml';
import 'ace-builds/src-noconflict/mode-json';
import 'ace-builds/src-noconflict/theme-github';
import 'ace-builds/src-noconflict/ext-language_tools';
import { projectApi } from '@/api/project';

const { Option } = Select;

interface ConfigEditorProps {
  projectId: number;
}

export default function ConfigEditor({ projectId }: ConfigEditorProps) {
  const [config, setConfig] = useState<string>('');
  const [format, setFormat] = useState<'yaml' | 'json'>('yaml');
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadConfig = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await projectApi.getConfig(projectId, format);
      
      // Response format: { data: "yaml or json string" }
      if (response.data && typeof response.data === 'string') {
        setConfig(response.data);
      } else if (response.data && typeof response.data === 'object') {
        // If it's an object, stringify it
        setConfig(JSON.stringify(response.data, null, 2));
      } else {
        setConfig('');
      }
    } catch (err: any) {
      const errorMsg = err?.response?.data?.error || err?.message || 'Failed to load config';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadConfig();
  }, [projectId, format]);

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    try {
      await projectApi.updateConfig(projectId, config, format);
      message.success('Config saved successfully');
      // Reload to get updated config
      await loadConfig();
    } catch (err: any) {
      const errorMsg = err?.response?.data?.error || err?.message || 'Failed to save config';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDownload = () => {
    const blob = new Blob([config], { 
      type: format === 'json' ? 'application/json' : 'application/x-yaml' 
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `project-${projectId}-config.${format === 'json' ? 'json' : 'yaml'}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    message.success('Config downloaded');
  };

  const handleFormatChange = (newFormat: 'yaml' | 'json') => {
    // When format changes, reload config in new format
    setFormat(newFormat);
    // Reload will happen in useEffect
  };

  if (isLoading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Spin size="large" />
        </div>
      </Card>
    );
  }

  return (
    <Card
      title={
        <Space>
          <span>Configuration Editor</span>
          <Select
            value={format}
            onChange={handleFormatChange}
            style={{ width: 100 }}
          >
            <Option value="yaml">YAML</Option>
            <Option value="json">JSON</Option>
          </Select>
        </Space>
      }
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={loadConfig}>
            Reload
          </Button>
          <Button icon={<DownloadOutlined />} onClick={handleDownload}>
            Download
          </Button>
          <Button
            type="primary"
            icon={<SaveOutlined />}
            onClick={handleSave}
            loading={isSaving}
          >
            Save
          </Button>
        </Space>
      }
    >
      {error && (
        <Alert
          message="Error"
          description={error}
          type="error"
          style={{ marginBottom: 16 }}
          closable
          onClose={() => setError(null)}
        />
      )}
      <div style={{ border: '1px solid #d9d9d9', borderRadius: '4px', overflow: 'hidden' }}>
        <AceEditor
          mode={format}
          theme="github"
          value={config}
          onChange={setConfig}
          width="100%"
          height="600px"
          fontSize={14}
          showPrintMargin={true}
          showGutter={true}
          highlightActiveLine={true}
          setOptions={{
            enableBasicAutocompletion: true,
            enableLiveAutocompletion: true,
            enableSnippets: true,
            showLineNumbers: true,
            tabSize: 2,
            useWorker: false,
          }}
        />
      </div>
      <div style={{ marginTop: 8, fontSize: '12px', color: '#888' }}>
        💡 Tip: Edit the configuration directly. Changes will be applied when you click Save.
      </div>
    </Card>
  );
}
