const { contextBridge, ipcRenderer } = require('electron');

// 暴露安全的API给渲染进程
contextBridge.exposeInMainWorld('electronAPI', {
  // 获取应用信息
  getAppInfo: () => {
    return {
      name: '小红书管理平台',
      version: '1.0.0',
      platform: process.platform
    };
  },

  // 显示消息框
  showMessage: (message, type = 'info') => {
    return ipcRenderer.invoke('show-message', message, type);
  },

  // 显示错误框
  showError: (title, message) => {
    return ipcRenderer.invoke('show-error', title, message);
  },

  // 选择文件
  selectFile: (options) => {
    return ipcRenderer.invoke('select-file', options);
  },

  // 选择文件夹
  selectFolder: () => {
    return ipcRenderer.invoke('select-folder');
  },

  // 打开外部链接
  openExternal: (url) => {
    return ipcRenderer.invoke('open-external', url);
  },

  // 获取后端服务状态
  getBackendStatus: () => {
    return ipcRenderer.invoke('get-backend-status');
  },

  // 重启后端服务
  restartBackend: () => {
    return ipcRenderer.invoke('restart-backend');
  }
});

// 监听来自主进程的消息
ipcRenderer.on('backend-status-changed', (event, status) => {
  // 通知渲染进程后端状态变化
  window.dispatchEvent(new CustomEvent('backend-status-changed', { detail: status }));
});

// 监听应用准备就绪
window.addEventListener('DOMContentLoaded', () => {
  console.log('Electron应用已准备就绪');
});
