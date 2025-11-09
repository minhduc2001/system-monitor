import { useState, useRef } from 'react';
import { Input, Button, Space, message, Modal, List, Tag, Card } from 'antd';
import type { MouseEvent } from 'react';
import { FolderOutlined, SearchOutlined, CheckOutlined } from '@ant-design/icons';

interface PathPickerProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  onServiceDetected?: (service: DetectedService) => void;
  onMultipleServicesDetected?: (services: DetectedService[]) => void;
}

export interface DetectedService {
  name: string;
  type: string;
  path: string;
  command: string;
  package_file: string;
}

export default function PathPicker({ value, onChange, placeholder, onServiceDetected, onMultipleServicesDetected }: PathPickerProps) {
  const [isDetecting, setIsDetecting] = useState(false);
  const [detectedServices, setDetectedServices] = useState<DetectedService[]>([]);
  const [showServices, setShowServices] = useState(false);
  const [showBrowseModal, setShowBrowseModal] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [parentPath, setParentPath] = useState<string>('');

  const handleBrowse = (e: MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setShowBrowseModal(true);
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    console.log('File input changed, files count:', files?.length || 0);
    
    if (files && files.length > 0) {
      // Get the first file to extract folder structure
      const file = files[0];
      // @ts-ignore - webkitRelativePath is available when using webkitdirectory
      const webkitPath = file.webkitRelativePath || '';
      
      console.log('First file webkitRelativePath:', webkitPath);
      
      if (webkitPath) {
        // Extract folder name from relative path
        // webkitRelativePath format: "folder-name/file.txt" or "folder-name/subfolder/file.txt"
        const pathParts = webkitPath.split('/');
        const folderName = pathParts[0]; // First part is usually the folder name
        
        console.log('Detected folder name:', folderName);
        
        // Try to detect if this is a project folder by checking for common files
        let foundPackageJson = false;
        let foundGoMod = false;
        let foundRequirementsTxt = false;
        let projectType = '';
        
        // Check first 100 files for performance
        const filesToCheck = Array.from(files).slice(0, 100);
        filesToCheck.forEach((f) => {
          const name = f.name.toLowerCase();
          // @ts-ignore
          const path = f.webkitRelativePath?.toLowerCase() || '';
          
          // Check root level files only (path has exactly 2 parts: folder/file)
          if (path.split('/').length === 2) {
            if (name === 'package.json') {
              foundPackageJson = true;
              projectType = 'Node.js/React/Vue';
            }
            if (name === 'go.mod') {
              foundGoMod = true;
              projectType = 'Go';
            }
            if (name === 'requirements.txt') {
              foundRequirementsTxt = true;
              projectType = 'Python';
            }
          }
        });
        
        // Determine OS
        const isWindows = window.navigator.platform.includes('Win') || window.navigator.userAgent.includes('Windows');
        const isMac = window.navigator.platform.includes('Mac') || window.navigator.userAgent.includes('Mac');
        const separator = isWindows ? '\\' : '/';
        
        // Smart path detection: Check if current value is a parent path
        let suggestedPath = '';
        const currentValue = value || '';
        
        console.log('Current value in input:', currentValue);
        console.log('Selected folder name:', folderName);
        
        // Smart path detection: Check if current value is a parent path
        // Strategy:
        // 1. If current value exists and doesn't contain the selected folder name, use it as parent
        // 2. If current value ends with / or \, definitely use it as parent
        // 3. If current value looks like an absolute path (starts with / or C:\), use it as parent
        // 4. Otherwise, create new path
        
        let isParentPath = false;
        
        if (currentValue) {
          const trimmedValue = currentValue.trim();
          
          // Check if current path already ends with this folder name (same folder selected)
          const endsWithFolder = trimmedValue.endsWith(folderName) || 
                                 trimmedValue.endsWith(folderName + '/') || 
                                 trimmedValue.endsWith(folderName + '\\');
          
          if (endsWithFolder) {
            // Same folder selected, keep current path
            suggestedPath = trimmedValue;
            console.log('Same folder selected, keeping path:', suggestedPath);
          } else if (trimmedValue.endsWith('/') || trimmedValue.endsWith('\\')) {
            // Ends with separator, definitely a parent path
            isParentPath = true;
          } else {
            // Check if it's likely a parent path
            // It's a parent if:
            // 1. It's an absolute path (starts with / or C:\)
            // 2. It doesn't contain the selected folder name
            // 3. It looks like a directory path (contains common path elements)
            const isAbsolutePath = trimmedValue.startsWith('/') || trimmedValue.match(/^[A-Z]:\\/);
            const containsFolderName = trimmedValue.includes(folderName);
            const looksLikeDirectory = trimmedValue.includes('Users') || 
                                       trimmedValue.includes('Documents') || 
                                       trimmedValue.includes('projects') || 
                                       trimmedValue.includes('home') ||
                                       trimmedValue.includes('Desktop');
            
            // If it's an absolute path that doesn't contain the folder name, it's likely a parent
            if (isAbsolutePath && !containsFolderName) {
              isParentPath = true;
            } 
            // Or if it looks like a directory and doesn't contain the folder name
            else if (looksLikeDirectory && !containsFolderName) {
              isParentPath = true;
            }
          }
        }
        
        if (isParentPath && currentValue) {
          // Use existing path as parent and append folder name
          const trimmedValue = currentValue.trim();
          if (trimmedValue.endsWith('/') || trimmedValue.endsWith('\\')) {
            suggestedPath = trimmedValue + folderName;
          } else {
            suggestedPath = trimmedValue + separator + folderName;
          }
          console.log('✅ Using existing parent path:', trimmedValue, '->', suggestedPath);
        } else if (!suggestedPath) {
          // No valid parent path, create new suggested path
          // Try to extract username from current value if it exists
          let usernameHint = 'YourUsername';
          if (currentValue) {
            // Try to extract username from current path (Unix/Mac format)
            const unixMatch = currentValue.match(/\/(Users|home)\/([^\/\\]+)/);
            // Try to extract username from Windows path
            const windowsMatch = currentValue.match(/[A-Z]:\\Users\\([^\\]+)/i);
            
            if (unixMatch && unixMatch[2] && unixMatch[2] !== 'YourUsername') {
              usernameHint = unixMatch[2];
            } else if (windowsMatch && windowsMatch[1] && windowsMatch[1] !== 'YourUsername') {
              usernameHint = windowsMatch[1];
            }
          }
          
          if (isWindows) {
            suggestedPath = `C:\\Users\\${usernameHint}\\Documents\\${folderName}`;
          } else if (isMac) {
            suggestedPath = `/Users/${usernameHint}/Documents/${folderName}`;
          } else {
            suggestedPath = `/home/${usernameHint}/Documents/${folderName}`;
          }
          console.log('Creating new suggested path:', suggestedPath);
        }
        
        // Fill in the input with suggested path
        console.log('Final suggested path:', suggestedPath);
        console.log('Was parent path used?', isParentPath);
        if (onChange) {
          onChange(suggestedPath);
        }
        
        // Show helpful message
        if (foundPackageJson || foundGoMod || foundRequirementsTxt) {
          message.success({
            content: (
              <div>
                <div>✅ Folder "<strong>{folderName}</strong>" selected successfully!</div>
                <div style={{ marginTop: 4 }}>
                  Project type detected: <strong>{projectType}</strong>
                </div>
                <div style={{ marginTop: 8, fontSize: '12px' }}>
                  Path has been filled in:
                  <br />
                  <code style={{ fontSize: '11px', background: '#f5f5f5', padding: '2px 4px', borderRadius: '2px', marginTop: '4px', display: 'inline-block' }}>
                    {suggestedPath}
                  </code>
                </div>
                {isParentPath && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#52c41a' }}>
                    ✅ Parent path detected and folder name appended automatically.
                  </div>
                )}
                {suggestedPath.includes('YourUsername') && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#fa8c16', background: '#fff7e6', padding: '8px', borderRadius: '4px', border: '1px solid #ffd591' }}>
                    <div><strong>⚠️ Important:</strong> Please replace "YourUsername" with your actual username!</div>
                    <div style={{ marginTop: 4 }}>
                      <strong>How to find your username:</strong>
                      <ul style={{ margin: '4px 0', paddingLeft: '20px' }}>
                        <li>On Mac/Linux: Open Terminal and run <code>whoami</code></li>
                        <li>On Windows: Check the path after <code>C:\Users\</code></li>
                        <li>Or check your home folder name in Finder/File Explorer</li>
                      </ul>
                    </div>
                    <div style={{ marginTop: 4 }}>
                      <strong>Example:</strong> If your username is "ducnm", change:
                      <br />
                      <code>/Users/YourUsername/Documents/...</code> → <code>/Users/ducnm/Documents/...</code>
                    </div>
                  </div>
                )}
                {!isParentPath && currentValue && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#faad14' }}>
                    💡 Tip: Enter the parent folder path first, then select subfolders to automatically combine paths.
                  </div>
                )}
                <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
                  💡 Verify the path is correct, then click "Detect" to auto-configure services.
                </div>
              </div>
            ),
            duration: 10,
          });
        } else {
          message.info({
            content: (
              <div>
                <div>📁 Folder "<strong>{folderName}</strong>" selected.</div>
                <div style={{ marginTop: 8, fontSize: '12px' }}>
                  Path has been filled in:
                  <br />
                  <code style={{ fontSize: '11px', background: '#f5f5f5', padding: '2px 4px', borderRadius: '2px', marginTop: '4px', display: 'inline-block' }}>
                    {suggestedPath}
                  </code>
                </div>
                {isParentPath && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#52c41a' }}>
                    ✅ Parent path detected and folder name appended automatically.
                  </div>
                )}
                {suggestedPath.includes('YourUsername') && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#fa8c16', background: '#fff7e6', padding: '8px', borderRadius: '4px', border: '1px solid #ffd591' }}>
                    <div><strong>⚠️ Important:</strong> Please replace "YourUsername" with your actual username!</div>
                    <div style={{ marginTop: 4 }}>
                      <strong>How to find your username:</strong>
                      <ul style={{ margin: '4px 0', paddingLeft: '20px' }}>
                        <li>On Mac/Linux: Open Terminal and run <code>whoami</code></li>
                        <li>On Windows: Check the path after <code>C:\Users\</code></li>
                        <li>Or check your home folder name in Finder/File Explorer</li>
                      </ul>
                    </div>
                    <div style={{ marginTop: 4 }}>
                      <strong>Example:</strong> If your username is "ducnm", change:
                      <br />
                      <code>/Users/YourUsername/Documents/...</code> → <code>/Users/ducnm/Documents/...</code>
                    </div>
                  </div>
                )}
                {!isParentPath && currentValue && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#faad14' }}>
                    💡 Tip: Enter the parent folder path first, then select subfolders to automatically combine paths.
                  </div>
                )}
                {!currentValue && (
                  <div style={{ marginTop: 8, fontSize: '12px', color: '#1890ff' }}>
                    💡 Tip: You can enter the parent folder path first, then select subfolders to automatically combine paths.
                  </div>
                )}
              </div>
            ),
            duration: 8,
          });
        }
      } else {
        // No webkitRelativePath - might be a single file or unsupported browser
        message.warning({
          content: 'Unable to detect folder structure. Please enter the full absolute path manually.',
          duration: 5,
        });
      }
    } else {
      console.log('No files selected');
    }
    
    // Reset input to allow selecting the same folder again
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };
  
  const handleSelectFolderClick = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    console.log('Select Folder button clicked');
    
    // Try File System Access API first (Chrome, Edge, Opera - requires HTTPS or localhost)
    // @ts-ignore
    const hasFileSystemAccess = window.showDirectoryPicker && 
      (window.location.protocol === 'https:' || window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');
    
    if (hasFileSystemAccess) {
      try {
        console.log('Using File System Access API');
        // @ts-ignore
        const dirHandle = await window.showDirectoryPicker();
        const folderName = dirHandle.name;
        
        console.log('Folder selected:', folderName);
        
        // Read directory to detect project type (limit to first 50 entries for performance)
        const entries: string[] = [];
        let count = 0;
        try {
          for await (const entry of dirHandle.values()) {
            if (count++ < 50) {
              entries.push(entry.name);
            } else {
              break;
            }
          }
        } catch (err) {
          console.warn('Error reading directory entries:', err);
        }
        
        const hasPackageJson = entries.includes('package.json');
        const hasGoMod = entries.includes('go.mod');
        const hasRequirementsTxt = entries.includes('requirements.txt');
        
        const projectType = hasPackageJson ? 'Node.js/React' : hasGoMod ? 'Go' : hasRequirementsTxt ? 'Python' : '';
        
        // Determine OS
        const isWindows = window.navigator.platform.includes('Win') || window.navigator.userAgent.includes('Windows');
        const isMac = window.navigator.platform.includes('Mac') || window.navigator.userAgent.includes('Mac');
        const separator = isWindows ? '\\' : '/';
        
        // Smart path detection for File System API
        let suggestedPath = '';
        const currentValue = value || '';
        
        // Check if current value looks like a valid parent directory path
        const isLikelyParentPath = currentValue && 
          (currentValue.endsWith('/') || currentValue.endsWith('\\') ||
           (!currentValue.includes(folderName) && 
            (currentValue.includes('Documents') || 
             currentValue.includes('projects') || 
             currentValue.includes('Users') || 
             currentValue.includes('home') ||
             currentValue.startsWith('/') ||
             currentValue.match(/^[A-Z]:\\/))));
        
        if (isLikelyParentPath) {
          // Use existing path as parent and append folder name
          if (currentValue.endsWith('/') || currentValue.endsWith('\\')) {
            suggestedPath = currentValue + folderName;
          } else {
            suggestedPath = currentValue + separator + folderName;
          }
          console.log('Using existing parent path (File System API):', currentValue, '->', suggestedPath);
        } else if (currentValue && currentValue.includes(folderName)) {
          // Path already contains this folder
          suggestedPath = currentValue;
          console.log('Path already contains folder, keeping:', suggestedPath);
        } else {
          // Create new suggested path
          // Try to extract username from current value if it exists
          let usernameHint = 'YourUsername';
          if (currentValue) {
            // Try to extract username from current path (Unix/Mac format)
            const unixMatch = currentValue.match(/\/(Users|home)\/([^\/\\]+)/);
            // Try to extract username from Windows path
            const windowsMatch = currentValue.match(/[A-Z]:\\Users\\([^\\]+)/i);
            
            if (unixMatch && unixMatch[2] && unixMatch[2] !== 'YourUsername') {
              usernameHint = unixMatch[2];
            } else if (windowsMatch && windowsMatch[1] && windowsMatch[1] !== 'YourUsername') {
              usernameHint = windowsMatch[1];
            }
          }
          
          if (isWindows) {
            suggestedPath = `C:\\Users\\${usernameHint}\\Documents\\${folderName}`;
          } else if (isMac) {
            suggestedPath = `/Users/${usernameHint}/Documents/${folderName}`;
          } else {
            suggestedPath = `/home/${usernameHint}/Documents/${folderName}`;
          }
          console.log('Creating new suggested path (File System API):', suggestedPath);
        }
        
        // Fill in the input with suggested path
        if (onChange) {
          onChange(suggestedPath);
        }
        
        message.success({
          content: (
            <div>
              <div>✅ Folder "<strong>{folderName}</strong>" selected.</div>
              {projectType && <div style={{ marginTop: 4 }}>Project type detected: <strong>{projectType}</strong></div>}
              <div style={{ marginTop: 8, fontSize: '12px' }}>
                Path has been filled in:
                <br />
                <code style={{ fontSize: '11px', background: '#f5f5f5', padding: '2px 4px', borderRadius: '2px', marginTop: '4px', display: 'inline-block' }}>
                  {suggestedPath}
                </code>
              </div>
              {suggestedPath.includes('YourUsername') && (
                <div style={{ marginTop: 8, fontSize: '12px', color: '#fa8c16', background: '#fff7e6', padding: '8px', borderRadius: '4px', border: '1px solid #ffd591' }}>
                  <div><strong>⚠️ Important:</strong> Please replace "YourUsername" with your actual username!</div>
                  <div style={{ marginTop: 4 }}>
                    <strong>How to find your username:</strong>
                    <ul style={{ margin: '4px 0', paddingLeft: '20px' }}>
                      <li>On Mac/Linux: Open Terminal and run <code>whoami</code></li>
                      <li>On Windows: Check the path after <code>C:\Users\</code></li>
                    </ul>
                  </div>
                </div>
              )}
              <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
                💡 Verify the path is correct, then click "Detect" to auto-configure services.
              </div>
            </div>
          ),
          duration: 10,
        });
        return;
      } catch (err: any) {
        if (err.name === 'AbortError' || err.name === 'NotAllowedError') {
          // User cancelled or denied permission
          console.log('User cancelled folder selection');
          return;
        } else {
          // Error occurred, fallback to traditional method
          console.error('File System Access API error:', err);
          // Fall through to traditional file input
        }
      }
    }
    
    // Fallback: Use traditional file input with webkitdirectory (works on all browsers and platforms)
    console.log('Using traditional file input method (webkitdirectory)');
    
    if (fileInputRef.current) {
      // Reset input to allow selecting the same folder again
      fileInputRef.current.value = '';
      // Trigger click programmatically
      fileInputRef.current.click();
      console.log('File input clicked programmatically');
    } else {
      console.error('File input ref not found');
      message.error('Unable to open folder picker. Please enter the path manually or check browser console.');
    }
  };

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    // Only set dragging to false if we're leaving the drop zone entirely
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const x = e.clientX;
    const y = e.clientY;
    if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
      setIsDragging(false);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const items = e.dataTransfer.items;
    if (items && items.length > 0) {
      // Check if folder is being dragged
      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        // @ts-ignore - webkitGetAsEntry is available in Chrome/Edge
        const entry = item.webkitGetAsEntry ? item.webkitGetAsEntry() : null;
        
        if (entry) {
          if (entry.isDirectory) {
            // Directory dropped
            const folderName = entry.name;
            
            // Try to read directory contents to detect project type
            try {
              // @ts-ignore
              const dirReader = entry.createReader();
              const entries: any[] = [];
              
              const readEntries = () => {
                return new Promise<void>((resolve) => {
                  // @ts-ignore
                  dirReader.readEntries((results: any[]) => {
                    if (results.length === 0) {
                      resolve();
                    } else {
                      entries.push(...results);
                      // Limit to first 50 entries
                      if (entries.length < 50) {
                        readEntries().then(resolve);
                      } else {
                        resolve();
                      }
                    }
                  });
                });
              };
              
              await readEntries();
              
              const fileNames = entries.map(e => e.name.toLowerCase());
              const hasPackageJson = fileNames.includes('package.json');
              const hasGoMod = fileNames.includes('go.mod');
              const hasRequirementsTxt = fileNames.includes('requirements.txt');
              
              const isWindows = window.navigator.platform.includes('Win') || window.navigator.userAgent.includes('Windows');
              const isMac = window.navigator.platform.includes('Mac') || window.navigator.userAgent.includes('Mac');
              const pathExample = isWindows 
                ? `C:\\Users\\username\\projects\\${folderName}`
                : isMac
                ? `/Users/username/projects/${folderName}`
                : `/home/username/projects/${folderName}`;
              
              if (hasPackageJson || hasGoMod || hasRequirementsTxt) {
                const projectType = hasPackageJson ? 'Node.js/React' : hasGoMod ? 'Go' : 'Python';
                message.success({
                  content: (
                    <div>
                      <div>✅ Folder "<strong>{folderName}</strong>" detected with project files!</div>
                      <div style={{ marginTop: 4 }}>Project type: <strong>{projectType}</strong></div>
                      <div style={{ marginTop: 8, fontSize: '12px' }}>
                        Please enter the full absolute path in the input field above:
                        <br />
                        <code style={{ fontSize: '11px', background: '#f5f5f5', padding: '2px 4px', borderRadius: '2px', marginTop: '4px', display: 'inline-block' }}>
                          {pathExample}
                        </code>
                      </div>
                      <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
                        💡 Then click "Detect" to auto-configure services.
                      </div>
                    </div>
                  ),
                  duration: 10,
                });
              } else {
                message.info({
                  content: (
                    <div>
                      <div>📁 Folder "<strong>{folderName}</strong>" detected.</div>
                      <div style={{ marginTop: 8, fontSize: '12px' }}>
                        Please enter the full absolute path:
                        <br />
                        <code style={{ fontSize: '11px', background: '#f5f5f5', padding: '2px 4px', borderRadius: '2px', marginTop: '4px', display: 'inline-block' }}>
                          {pathExample}
                        </code>
                      </div>
                    </div>
                  ),
                  duration: 8,
                });
              }
            } catch (err) {
              message.info({
                content: `📁 Folder "${entry.name}" detected. Please enter the full absolute path manually.`,
                duration: 5,
              });
            }
            return;
          } else if (entry.isFile) {
            // File dropped - try to get folder path
            // @ts-ignore
            entry.file((file: File) => {
              // @ts-ignore
              const webkitPath = file.webkitRelativePath || '';
              if (webkitPath) {
                const pathParts = webkitPath.split('/');
                if (pathParts.length > 1) {
                  const folderName = pathParts[0];
                  message.info({
                    content: `File from folder "${folderName}" detected. Please enter the full absolute path of the folder.`,
                    duration: 5,
                  });
                }
              }
            });
            return;
          }
        }
      }
    }

    // Fallback: show instructions
    message.info({
      content: 'Please use the "Select Folder" button or enter the path manually. Drag & drop works best with the file picker.',
      duration: 4,
    });
  };

  const handleDetectServices = async () => {
    if (!value || value.trim() === '') {
      message.warning('Please enter a project path first');
      return;
    }

    const pathValue = value.trim();
    
    // Check if path contains placeholder "YourUsername"
    if (pathValue.includes('YourUsername')) {
      message.error({
        content: (
          <div>
            <div><strong>Invalid path:</strong> The path contains "YourUsername" placeholder.</div>
            <div style={{ marginTop: 8, fontSize: '12px' }}>
              Please replace <code>YourUsername</code> with your actual username.
              <br />
              <strong>Example:</strong> If your path is <code>/Users/YourUsername/Documents/da-ban-quan-ao-ai</code>
              <br />
              Replace it with: <code>/Users/ducnm/Documents/da-ban-quan-ao-ai</code> (use your actual username)
            </div>
            <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
              💡 <strong>Tip:</strong> On Mac, you can find your username by running <code>whoami</code> in Terminal.
            </div>
          </div>
        ),
        duration: 10,
      });
      return;
    }

    setIsDetecting(true);
    try {
      const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
      const response = await fetch(`${baseURL}/api/v1/projects/detect-services`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ path: pathValue }),
      });

      const data = await response.json();
      if (response.ok && data.data) {
        setDetectedServices(data.data);
        if (data.data.length > 0) {
          setShowServices(true);
          message.success(`Found ${data.data.length} service(s) in this path`);
          // Notify parent if there are multiple services
          if (data.data.length > 1 && onMultipleServicesDetected) {
            onMultipleServicesDetected(data.data);
          }
        } else {
          message.info('No services detected in this path');
        }
      } else {
        // Enhanced error message for path not found
        const errorMsg = data.error || 'Failed to detect services';
        if (errorMsg.includes('Path does not exist') || errorMsg.includes('does not exist')) {
          message.error({
            content: (
              <div>
                <div><strong>Path does not exist:</strong> <code>{pathValue}</code></div>
                <div style={{ marginTop: 8, fontSize: '12px' }}>
                  Please verify:
                  <ul style={{ margin: '4px 0', paddingLeft: '20px' }}>
                    <li>The path is correct and the folder exists</li>
                    <li>If you used "Select Folder", make sure to replace "YourUsername" with your actual username</li>
                    <li>On Mac: Check your username with <code>whoami</code> in Terminal</li>
                    <li>On Windows: Check your username in the path (usually after <code>C:\Users\</code>)</li>
                  </ul>
                </div>
              </div>
            ),
            duration: 10,
          });
        } else {
          message.error(errorMsg);
        }
      }
    } catch (error) {
      message.error('Failed to detect services. Please check your connection.');
    } finally {
      setIsDetecting(false);
    }
  };

  const handleSelectService = (service: DetectedService) => {
    if (onChange) {
      onChange(service.path);
    }
    // Notify parent component about the selected service
    if (onServiceDetected) {
      onServiceDetected(service);
    }
    setShowServices(false);
    message.success(`Selected: ${service.name}`);
  };

  return (
    <div>
      <div
        onDragEnter={handleDragEnter}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        style={{
          position: 'relative',
          border: isDragging ? '2px dashed #1890ff' : '2px dashed transparent',
          borderRadius: '4px',
          padding: isDragging ? '2px' : '0',
          transition: 'all 0.3s',
        }}
      >
        <Input.Group compact style={{ display: 'flex' }}>
          <Input
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            placeholder={placeholder || 'Enter project path, drag folder here, or click browse'}
            style={{ flex: 1 }}
            onDragEnter={handleDragEnter}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          />
          <Button
            type="default"
            icon={<SearchOutlined />}
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              handleDetectServices();
            }}
            loading={isDetecting}
            title="Detect services in this path"
          >
            Detect
          </Button>
          <Button
            type="default"
            icon={<FolderOutlined />}
            onClick={handleBrowse}
            title="Browse folder (shows instructions)"
          >
            Browse
          </Button>
          <input
            type="file"
            ref={fileInputRef}
            {...({ webkitdirectory: '' } as any)}
            {...({ directory: '' } as any)}
            multiple
            style={{ display: 'none' }}
            id="folder-input"
            onChange={handleFileInputChange}
          />
          <Button
            type="primary"
            icon={<FolderOutlined />}
            onClick={handleSelectFolderClick}
            title="Select folder - Works on Mac, Windows, and Linux"
          >
            Select Folder
          </Button>
        </Input.Group>
        {isDragging && (
          <div
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'rgba(24, 144, 255, 0.1)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: '4px',
              zIndex: 10,
              pointerEvents: 'none',
            }}
          >
            <div style={{ textAlign: 'center', color: '#1890ff', fontWeight: 'bold' }}>
              Drop folder here
            </div>
          </div>
        )}
      </div>

      <Modal
        title="Select Project Folder"
        open={showBrowseModal}
        onCancel={() => setShowBrowseModal(false)}
        onOk={() => setShowBrowseModal(false)}
        width={600}
      >
        <div>
          <p>Browser security restrictions prevent direct folder selection.</p>
          <p><strong>How to get folder path:</strong></p>
          <ol>
            <li><strong>On macOS:</strong>
              <ul>
                <li>Open Finder and navigate to your project folder</li>
                <li>Right-click the folder and select "Get Info"</li>
                <li>Copy the path from "Where" field, or</li>
                <li>Drag and drop the folder into Terminal to see its path</li>
                <li>Or press Cmd+Option+C in Finder to copy the path</li>
              </ul>
            </li>
            <li><strong>On Windows:</strong>
              <ul>
                <li>Open File Explorer and navigate to your project folder</li>
                <li>Click on the address bar to see the full path</li>
                <li>Or Shift+Right-click the folder and select "Copy as path"</li>
              </ul>
            </li>
            <li><strong>On Linux:</strong>
              <ul>
                <li>Right-click the folder and select "Properties" to see the path</li>
                <li>Or use the file manager's address bar</li>
              </ul>
            </li>
          </ol>
          <p><strong>Then paste the path in the input field above.</strong></p>
        </div>
      </Modal>

      {showServices && detectedServices.length > 0 && (
        <Card
          title={`Detected Services (${detectedServices.length})`}
          style={{ marginTop: 16 }}
          extra={
            <Button type="link" onClick={() => setShowServices(false)}>
              Close
            </Button>
          }
        >
          <List
            dataSource={detectedServices}
            renderItem={(service) => (
              <List.Item
                actions={[
                  <Button
                    type="primary"
                    icon={<CheckOutlined />}
                    onClick={() => handleSelectService(service)}
                  >
                    Use This
                  </Button>,
                ]}
              >
                <List.Item.Meta
                  title={
                    <Space>
                      <span>{service.name}</span>
                      <Tag color={service.type === 'frontend' ? 'green' : service.type === 'backend' ? 'blue' : 'default'}>
                        {service.type}
                      </Tag>
                    </Space>
                  }
                  description={
                    <div>
                      <div style={{ marginBottom: 4 }}>
                        <strong>Path:</strong> <code style={{ fontSize: '12px' }}>{service.path}</code>
                      </div>
                      {service.command && (
                        <div>
                          <strong>Command:</strong> <code style={{ fontSize: '12px' }}>{service.command}</code>
                        </div>
                      )}
                    </div>
                  }
                />
              </List.Item>
            )}
          />
          <div style={{ marginTop: 16, fontSize: '12px', color: '#888' }}>
            💡 Select a service to auto-fill the form. If you have multiple services (FE, BE, etc.), 
            you can create separate projects for each one by selecting them one by one.
          </div>
        </Card>
      )}
    </div>
  );
}
