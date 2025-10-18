# Debug Guide

## 🐛 Common Errors & Solutions

### 1. TypeError: uptime.match is not a function

**Vấn đề**: API trả về `uptime` không phải là string, có thể là number hoặc object.

**Giải pháp**:

- Cập nhật type definition: `uptime: string | number`
- Sử dụng try-catch trong `formatUptime` function
- Kiểm tra type trước khi gọi `.match()`

**Code fix**:

```typescript
export function formatUptime(
  uptime: string | number | undefined | null
): string {
  try {
    if (!uptime) return "0s";

    // Convert to string if it's a number
    const uptimeStr = typeof uptime === "number" ? uptime.toString() : uptime;

    // Check if it's a string and has match method
    if (typeof uptimeStr === "string" && uptimeStr.match) {
      // Safe to call .match()
    }

    return "0s";
  } catch (error) {
    console.warn("Error formatting uptime:", error);
    return "0s";
  }
}
```

### 2. Ant Design Icons Missing

**Vấn đề**: Một số icons không có trong `@ant-design/icons`.

**Giải pháp**: Sử dụng `lucide-react` thay thế.

**Code fix**:

```typescript
// Before
import { CpuOutlined } from "@ant-design/icons";

// After
import { CpuIcon } from "lucide-react";
```

### 3. TypeScript Strict Mode Errors

**Vấn đề**: `interface` là reserved word trong strict mode.

**Giải pháp**: Đổi tên parameter.

**Code fix**:

```typescript
// Before
export function getNetworkInterfaceStatus(interface: any): string;

// After
export function getNetworkInterfaceStatus(networkInterface: any): string;
```

### 4. Unused Imports

**Vấn đề**: ESLint báo unused imports.

**Giải pháp**: Xóa imports không sử dụng.

**Code fix**:

```typescript
// Before
import React, { useState } from "react";
import { Card, Row, Col } from "antd";

// After (nếu không dùng React và Row, Col)
import { useState } from "react";
import { Card } from "antd";
```

## 🔍 Debugging Tips

### 1. Check API Response

```typescript
console.log("API Response:", data);
console.log("Uptime type:", typeof data.uptime);
console.log("Uptime value:", data.uptime);
```

### 2. Add Error Boundaries

```typescript
try {
  // Risky code
} catch (error) {
  console.error("Error:", error);
  // Fallback
}
```

### 3. Type Guards

```typescript
function isString(value: any): value is string {
  return typeof value === "string";
}

if (isString(uptime)) {
  // Safe to use string methods
}
```

### 4. Default Values

```typescript
const safeUptime = uptime || "0s";
const safeValue = value ?? defaultValue;
```

## 🛠️ Development Tools

### 1. Browser DevTools

- Console để xem errors
- Network tab để check API calls
- React DevTools để debug components

### 2. VS Code Extensions

- TypeScript Hero
- ESLint
- Prettier
- Auto Rename Tag

### 3. Debug Commands

```bash
# Check TypeScript errors
npm run type-check

# Check linting errors
npm run lint

# Build and check for errors
npm run build
```

## 📋 Checklist

### Before Deploy

- [ ] No TypeScript errors
- [ ] No ESLint warnings
- [ ] All tests pass
- [ ] Build successful
- [ ] API integration works
- [ ] Error handling in place

### Common Issues

- [ ] Check API response format
- [ ] Verify type definitions
- [ ] Test with different data types
- [ ] Add fallback values
- [ ] Handle null/undefined cases

## 🚀 Quick Fixes

### 1. Safe Property Access

```typescript
const value = data?.property?.subProperty || "default";
```

### 2. Type Assertion

```typescript
const stringValue = value as string;
```

### 3. Optional Chaining

```typescript
const result = data?.method?.();
```

### 4. Nullish Coalescing

```typescript
const result = value ?? "default";
```
