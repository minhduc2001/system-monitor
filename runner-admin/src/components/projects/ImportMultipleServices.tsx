import { useState, useMemo } from 'react';
import { Modal, List, Tag, Button, Space, message, Checkbox, Input, Empty } from 'antd';
import { PlusOutlined, SearchOutlined } from '@ant-design/icons';
import type { DetectedService } from './PathPicker';
import { useCreateProject } from '@/hooks/queries/use-project.query';
import type { CreateProjectRequest } from '@/types/project';

interface ImportMultipleServicesProps {
  services: DetectedService[];
  visible: boolean;
  onClose: () => void;
  onComplete: () => void;
}

export default function ImportMultipleServices({
  services,
  visible,
  onClose,
  onComplete,
}: ImportMultipleServicesProps) {
  const [selectedServices, setSelectedServices] = useState<number[]>([]);
  const [searchText, setSearchText] = useState('');
  const createProject = useCreateProject();

  // Filter services based on search text
  const filteredServices = useMemo(() => {
    if (!searchText.trim()) {
      return services;
    }
    const search = searchText.toLowerCase();
    return services.filter(
      (service) =>
        service.name.toLowerCase().includes(search) ||
        service.path.toLowerCase().includes(search) ||
        service.type.toLowerCase().includes(search)
    );
  }, [services, searchText]);

  const handleToggleService = (serviceIndex: number) => {
    // Find the actual index in the original services array
    const service = filteredServices[serviceIndex];
    const originalIndex = services.findIndex((s) => s.path === service.path && s.name === service.name);
    
    if (originalIndex === -1) return;
    
    if (selectedServices.includes(originalIndex)) {
      setSelectedServices(selectedServices.filter((i) => i !== originalIndex));
    } else {
      setSelectedServices([...selectedServices, originalIndex]);
    }
  };

  const handleImportSelected = async () => {
    if (selectedServices.length === 0) {
      message.warning('Please select at least one service to import');
      return;
    }

    const servicesToImport = selectedServices.map((index) => services[index]);
    let successCount = 0;
    let errorCount = 0;

    for (const service of servicesToImport) {
      try {
        const projectData: CreateProjectRequest = {
          name: service.name,
          type: service.type as any,
          path: service.path,
          command: service.command,
          environment: 'development',
          auto_restart: true,
        };
        await createProject.mutateAsync(projectData);
        successCount++;
      } catch (error) {
        errorCount++;
        console.error(`Failed to create project for ${service.name}:`, error);
      }
    }

    if (successCount > 0) {
      message.success(`Successfully imported ${successCount} service(s)`);
      onComplete();
    }
    if (errorCount > 0) {
      message.warning(`Failed to import ${errorCount} service(s). Check console for details.`);
    }

    onClose();
    setSelectedServices([]);
  };

  // Check if all filtered services are selected
  const allFilteredSelected = useMemo(() => {
    if (filteredServices.length === 0) return false;
    return filteredServices.every((service) => {
      const originalIndex = services.findIndex((s) => s.path === service.path && s.name === service.name);
      return originalIndex !== -1 && selectedServices.includes(originalIndex);
    });
  }, [filteredServices, services, selectedServices]);

  // Check if some filtered services are selected
  const someFilteredSelected = useMemo(() => {
    if (filteredServices.length === 0) return false;
    const selectedCount = filteredServices.filter((service) => {
      const originalIndex = services.findIndex((s) => s.path === service.path && s.name === service.name);
      return originalIndex !== -1 && selectedServices.includes(originalIndex);
    }).length;
    return selectedCount > 0 && selectedCount < filteredServices.length;
  }, [filteredServices, services, selectedServices]);

  const handleSelectAllFiltered = (checked: boolean) => {
    if (checked) {
      // Select all filtered services
      const newSelected = new Set(selectedServices);
      filteredServices.forEach((service) => {
        const originalIndex = services.findIndex((s) => s.path === service.path && s.name === service.name);
        if (originalIndex !== -1) {
          newSelected.add(originalIndex);
        }
      });
      setSelectedServices(Array.from(newSelected));
    } else {
      // Deselect all filtered services
      const filteredIndices = new Set(
        filteredServices.map((service) => {
          return services.findIndex((s) => s.path === service.path && s.name === service.name);
        }).filter((idx) => idx !== -1)
      );
      setSelectedServices(selectedServices.filter((idx) => !filteredIndices.has(idx)));
    }
  };

  return (
    <Modal
      title={`Import Multiple Services (${services.length} found${filteredServices.length !== services.length ? `, ${filteredServices.length} shown` : ''})`}
      open={visible}
      onCancel={() => {
        onClose();
        setSearchText('');
        setSelectedServices([]);
      }}
      width={900}
      style={{ top: 20 }}
      footer={[
        <Button key="cancel" onClick={() => {
          onClose();
          setSearchText('');
          setSelectedServices([]);
        }}>
          Cancel
        </Button>,
        <Button
          key="import"
          type="primary"
          icon={<PlusOutlined />}
          onClick={handleImportSelected}
          loading={createProject.isPending}
          disabled={selectedServices.length === 0}
        >
          Import Selected ({selectedServices.length})
        </Button>,
      ]}
    >
      <div style={{ marginBottom: 16 }}>
        <Input
          placeholder="Search services by name, path, or type..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          allowClear
          style={{ marginBottom: 12 }}
        />
        <Checkbox
          checked={allFilteredSelected}
          indeterminate={someFilteredSelected}
          onChange={(e) => handleSelectAllFiltered(e.target.checked)}
        >
          Select All{filteredServices.length !== services.length ? ` (${filteredServices.length} shown)` : ''}
        </Checkbox>
        {selectedServices.length > 0 && (
          <span style={{ marginLeft: 16, color: '#666', fontSize: '12px' }}>
            {selectedServices.length} service(s) selected
          </span>
        )}
      </div>
      <div style={{ maxHeight: '60vh', overflowY: 'auto', border: '1px solid #f0f0f0', borderRadius: '4px', padding: '8px' }}>
        {filteredServices.length === 0 ? (
          <Empty description="No services found" style={{ padding: '40px 0' }} />
        ) : (
          <List
            dataSource={filteredServices}
            renderItem={(service, index) => {
              const originalIndex = services.findIndex((s) => s.path === service.path && s.name === service.name);
              const isSelected = originalIndex !== -1 && selectedServices.includes(originalIndex);
              
              return (
                <List.Item
                  style={{
                    padding: '12px',
                    borderBottom: '1px solid #f0f0f0',
                    cursor: 'pointer',
                    backgroundColor: isSelected ? '#e6f7ff' : 'transparent',
                  }}
                  onClick={() => handleToggleService(index)}
                >
                  <div style={{ display: 'flex', alignItems: 'flex-start', width: '100%' }}>
                    <Checkbox
                      checked={isSelected}
                      onChange={() => handleToggleService(index)}
                      onClick={(e) => e.stopPropagation()}
                      style={{ marginRight: '12px', marginTop: '4px' }}
                    />
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ marginBottom: '8px' }}>
                        <Space>
                          <span style={{ fontWeight: 500 }}>{service.name}</span>
                          <Tag
                            color={
                              service.type === 'frontend'
                                ? 'green'
                                : service.type === 'backend'
                                ? 'blue'
                                : 'default'
                            }
                          >
                            {service.type}
                          </Tag>
                        </Space>
                      </div>
                      <div style={{ fontSize: '12px', color: '#666', lineHeight: '1.6' }}>
                        <div style={{ marginBottom: '4px' }}>
                          <strong style={{ marginRight: '4px' }}>Path:</strong>
                          <code
                            style={{
                              fontSize: '11px',
                              background: '#f5f5f5',
                              padding: '2px 6px',
                              borderRadius: '3px',
                              fontFamily: 'Monaco, Menlo, "Courier New", monospace',
                              display: 'inline-block',
                              maxWidth: '100%',
                              whiteSpace: 'nowrap',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                            }}
                            title={service.path}
                          >
                            {service.path}
                          </code>
                        </div>
                        {service.command && (
                          <div>
                            <strong style={{ marginRight: '4px' }}>Command:</strong>
                            <code
                              style={{
                                fontSize: '11px',
                                background: '#f5f5f5',
                                padding: '2px 6px',
                                borderRadius: '3px',
                                fontFamily: 'Monaco, Menlo, "Courier New", monospace',
                                display: 'inline-block',
                                maxWidth: '100%',
                                whiteSpace: 'nowrap',
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                              }}
                              title={service.command}
                            >
                              {service.command}
                            </code>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </List.Item>
              );
            }}
          />
        )}
      </div>
      {services.length > 50 && (
        <div style={{ marginTop: 12, fontSize: '12px', color: '#999', textAlign: 'center' }}>
          💡 Tip: Use the search box to filter services. Showing {filteredServices.length} of {services.length} services.
        </div>
      )}
    </Modal>
  );
}

