const { app, BrowserWindow, Menu, dialog, shell } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');

// 后端服务进程
let backendProcess = null;
let BACKEND_PORT = 18060;

// 创建主窗口
function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      enableRemoteModule: false,
      preload: path.join(__dirname, 'preload.js')
    },
    icon: path.join(__dirname, 'assets/icon.png'),
    title: '小红书管理平台',
    show: false // 先不显示，等后端启动后再显示
  });

  // 设置菜单
  createMenu(mainWindow);

  // 启动后端服务
  startBackendService().then(() => {
    // 后端启动成功后加载页面
    mainWindow.loadURL(`http://localhost:${BACKEND_PORT}`);
    mainWindow.show();
    
    // 开发模式下打开开发者工具
    if (process.env.NODE_ENV === 'development') {
      mainWindow.webContents.openDevTools();
    }
  }).catch((error) => {
    console.error('后端服务启动失败:', error);
    
    let errorMessage = '无法启动后端服务\n\n';
    if (error.message.includes('端口被占用')) {
      errorMessage += '错误原因：端口被占用\n\n';
      errorMessage += '解决方案：\n';
      errorMessage += '1. 关闭其他可能占用端口的程序\n';
      errorMessage += '2. 重启应用\n';
      errorMessage += '3. 运行 kill-port.bat 清理端口\n';
      errorMessage += '4. 重启计算机';
    } else if (error.message.includes('后端可执行文件不存在')) {
      errorMessage += '错误原因：后端程序文件缺失\n\n';
      errorMessage += '解决方案：\n';
      errorMessage += '1. 重新运行 build-electron.bat 构建应用\n';
      errorMessage += '2. 确保 Eapp/backend/ 目录下有 xiaohongshu-mcp-desktop.exe 文件';
    } else if (error.message.includes('无法找到可用端口')) {
      errorMessage += '错误原因：所有端口都被占用\n\n';
      errorMessage += '解决方案：\n';
      errorMessage += '1. 关闭其他占用端口的程序\n';
      errorMessage += '2. 重启计算机\n';
      errorMessage += '3. 检查防火墙设置';
    } else {
      errorMessage += `错误详情：${error.message}\n\n`;
      errorMessage += '解决方案：\n';
      errorMessage += '1. 重启应用\n';
      errorMessage += '2. 检查系统资源\n';
      errorMessage += '3. 查看控制台日志获取更多信息';
    }
    
    dialog.showErrorBox('启动失败', errorMessage);
    app.quit();
  });

  // 处理窗口关闭
  mainWindow.on('closed', () => {
    stopBackendService();
  });

  // 处理外部链接
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  return mainWindow;
}

// 创建应用菜单
function createMenu(mainWindow) {
  const template = [
    {
      label: '文件',
      submenu: [
        {
          label: '刷新',
          accelerator: 'CmdOrCtrl+R',
          click: () => {
            mainWindow.reload();
          }
        },
        {
          label: '退出',
          accelerator: process.platform === 'darwin' ? 'Cmd+Q' : 'Ctrl+Q',
          click: () => {
            app.quit();
          }
        }
      ]
    },
    {
      label: '编辑',
      submenu: [
        { role: 'undo', label: '撤销' },
        { role: 'redo', label: '重做' },
        { type: 'separator' },
        { role: 'cut', label: '剪切' },
        { role: 'copy', label: '复制' },
        { role: 'paste', label: '粘贴' }
      ]
    },
    {
      label: '视图',
      submenu: [
        { role: 'reload', label: '重新加载' },
        { role: 'forceReload', label: '强制重新加载' },
        { role: 'toggleDevTools', label: '开发者工具' },
        { type: 'separator' },
        { role: 'resetZoom', label: '实际大小' },
        { role: 'zoomIn', label: '放大' },
        { role: 'zoomOut', label: '缩小' },
        { type: 'separator' },
        { role: 'togglefullscreen', label: '全屏' }
      ]
    },
    {
      label: '帮助',
      submenu: [
        {
          label: '关于',
          click: () => {
            dialog.showMessageBox(mainWindow, {
              type: 'info',
              title: '关于',
              message: '小红书管理平台',
              detail: '版本 1.0.0\n一个功能强大的小红书内容管理工具'
            });
          }
        }
      ]
    }
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}

// 检查端口是否可用
function checkPortAvailable(port) {
  return new Promise((resolve) => {
    const net = require('net');
    const server = net.createServer();
    
    server.listen(port, () => {
      server.once('close', () => {
        resolve(true);
      });
      server.close();
    });
    
    server.on('error', (err) => {
      console.log(`端口 ${port} 不可用:`, err.message);
      resolve(false);
    });
  });
}

// 查找可用端口
async function findAvailablePort(startPort = 18060) {
  console.log(`开始查找可用端口，起始端口: ${startPort}`);
  
  for (let port = startPort; port < startPort + 100; port++) {
    console.log(`检查端口 ${port}...`);
    if (await checkPortAvailable(port)) {
      console.log(`找到可用端口: ${port}`);
      return port;
    }
  }
  
  // 如果18060-18159都被占用，尝试其他端口范围
  console.log('端口18060-18159都被占用，尝试其他端口范围...');
  for (let port = 8080; port < 8180; port++) {
    console.log(`检查端口 ${port}...`);
    if (await checkPortAvailable(port)) {
      console.log(`找到可用端口: ${port}`);
      return port;
    }
  }
  
  throw new Error('无法找到可用端口，请检查是否有其他程序占用了大量端口');
}

// 启动后端服务
async function startBackendService() {
  return new Promise(async (resolve, reject) => {
    try {
      // 检查后端可执行文件是否存在
      // 在开发环境中，__dirname指向Eapp目录
      // 在打包后的应用中，__dirname指向resources/app.asar.unpacked目录
      let backendPath;
      
      // 尝试多个可能的路径，按优先级排序
      const possiblePaths = [];
      
      // 如果是打包后的应用，优先使用可执行的路径
      if (app.isPackaged && process.resourcesPath) {
        possiblePaths.push(
          path.join(process.resourcesPath, 'app.asar.unpacked', 'backend', 'xiaohongshu-mcp-desktop.exe'), // 最高优先级：asar.unpacked
          path.join(process.resourcesPath, 'backend', 'xiaohongshu-mcp-desktop.exe'), // 次优先级：resources/backend
          path.join(process.resourcesPath, 'app', 'backend', 'xiaohongshu-mcp-desktop.exe'),
          path.join(process.resourcesPath, '..', 'backend', 'xiaohongshu-mcp-desktop.exe'),
          path.join(process.resourcesPath, 'app.asar.unpacked', 'backend', 'xiaohongshu-mcp.exe'), // 最高优先级：asar.unpacked
          path.join(process.resourcesPath, 'backend', 'xiaohongshu-mcp.exe'), // 次优先级：resources/backend
          path.join(process.resourcesPath, 'app', 'backend', 'xiaohongshu-mcp.exe'),
          path.join(process.resourcesPath, '..', 'backend', 'xiaohongshu-mcp.exe')
        );
      }
      
      // 开发环境路径
      possiblePaths.push(
        path.join(__dirname, 'backend', 'xiaohongshu-mcp-desktop.exe'), // 开发环境
        path.join(__dirname, '..', 'backend', 'xiaohongshu-mcp-desktop.exe'), // 备用路径1
        path.join(__dirname, '..', '..', 'backend', 'xiaohongshu-mcp-desktop.exe'), // 备用路径2
        path.join(process.cwd(), 'backend', 'xiaohongshu-mcp-desktop.exe'), // 当前工作目录
        path.join(process.cwd(), '..', 'Eapp', 'backend', 'xiaohongshu-mcp-desktop.exe') // 项目根目录
      );
      
      // 如果是打包后的应用，尝试更多路径
      if (app.isPackaged) {
        const appPath = app.getAppPath();
        const exePath = process.execPath;
        const exeDir = path.dirname(exePath);
        
        possiblePaths.push(
          path.join(appPath, 'backend', 'xiaohongshu-mcp-desktop.exe'),
          path.join(exeDir, 'backend', 'xiaohongshu-mcp-desktop.exe'),
          path.join(exeDir, '..', 'backend', 'xiaohongshu-mcp-desktop.exe'),
          path.join(exeDir, '..', '..', 'backend', 'xiaohongshu-mcp-desktop.exe')
        );
      }
      
      // 查找存在的后端文件
      console.log('=== 后端文件路径检测 ===');
      console.log('__dirname:', __dirname);
      console.log('process.cwd():', process.cwd());
      console.log('process.resourcesPath:', process.resourcesPath);
      console.log('process.execPath:', process.execPath);
      console.log('app.isPackaged:', app.isPackaged);
      if (app.isPackaged) {
        console.log('app.getAppPath():', app.getAppPath());
        console.log('process.execPath dirname:', path.dirname(process.execPath));
      }
      console.log('');
      
      console.log('尝试的路径:');
      possiblePaths.forEach((p, index) => {
        const exists = fs.existsSync(p);
        const isExecutable = exists && !p.includes('app.asar') && !p.includes('app\\');
        console.log(`${index + 1}. ${p} - ${exists ? '✓ 存在' : '✗ 不存在'} ${isExecutable ? '(可执行)' : '(不可执行)'}`);
      });
      
      // 优先选择可执行的路径
      backendPath = possiblePaths.find(p => {
        const exists = fs.existsSync(p);
        const isExecutable = !p.includes('app.asar') && !p.includes('app\\');
        return exists && isExecutable;
      });
      
      // 如果没有找到可执行的路径，使用第一个存在的路径
      if (!backendPath) {
        backendPath = possiblePaths.find(p => fs.existsSync(p));
      }
      
      if (!backendPath) {
        console.log('\n❌ 未找到后端文件');
        reject(new Error('后端可执行文件不存在，已尝试多个路径'));
        return;
      }
      
      console.log('\n✅ 找到后端文件:', backendPath);

      // 查找可用端口
      BACKEND_PORT = await findAvailablePort(18060);
      console.log('使用端口:', BACKEND_PORT);

      console.log('启动后端服务:', backendPath);
      
      // 启动后端进程
      backendProcess = spawn(backendPath, ['-no-browser', '-port', BACKEND_PORT.toString()], {
        cwd: path.dirname(backendPath),
        stdio: ['ignore', 'pipe', 'pipe']
      });

      // 处理后端进程输出
      backendProcess.stdout.on('data', (data) => {
        const output = data.toString();
        console.log('后端输出:', output);
        
        // 检查是否有端口占用错误
        if (output.includes('端口被占用') || output.includes('port already in use') || output.includes('bind: address already in use')) {
          console.error('检测到端口占用错误');
          reject(new Error('端口被占用，请关闭其他占用端口的程序或重启应用'));
        }
      });

      backendProcess.stderr.on('data', (data) => {
        const error = data.toString();
        console.error('后端错误:', error);
        
        // 检查是否有端口占用错误
        if (error.includes('端口被占用') || error.includes('port already in use') || error.includes('bind: address already in use')) {
          console.error('检测到端口占用错误');
          reject(new Error('端口被占用，请关闭其他占用端口的程序或重启应用'));
        }
      });

      backendProcess.on('error', (error) => {
        console.error('后端进程错误:', error);
        reject(error);
      });

      backendProcess.on('exit', (code) => {
        console.log('后端进程退出，代码:', code);
        if (code !== 0) {
          reject(new Error(`后端进程异常退出，代码: ${code}`));
        }
      });

      // 等待后端服务启动，增加重试机制
      let retryCount = 0;
      const maxRetries = 5;
      
      const checkService = async () => {
        try {
          await checkBackendHealth();
          console.log('后端服务启动成功');
          resolve();
        } catch (error) {
          retryCount++;
          console.log(`后端服务健康检查失败 (${retryCount}/${maxRetries}):`, error.message);
          
          if (retryCount >= maxRetries) {
            reject(new Error(`后端服务启动失败，已重试 ${maxRetries} 次`));
          } else {
            // 等待2秒后重试
            setTimeout(checkService, 2000);
          }
        }
      };
      
      // 初始等待3秒
      setTimeout(checkService, 3000);
    } catch (error) {
      reject(error);
    }
  });
}

// 检查后端服务健康状态
function checkBackendHealth() {
  return new Promise((resolve, reject) => {
    const http = require('http');
    const options = {
      hostname: 'localhost',
      port: BACKEND_PORT,
      path: '/health',
      method: 'GET',
      timeout: 5000
    };

    const req = http.request(options, (res) => {
      if (res.statusCode === 200) {
        resolve();
      } else {
        reject(new Error(`健康检查失败，状态码: ${res.statusCode}`));
      }
    });

    req.on('error', (error) => {
      reject(error);
    });

    req.on('timeout', () => {
      req.destroy();
      reject(new Error('健康检查超时'));
    });

    req.end();
  });
}

// 停止后端服务
function stopBackendService() {
  if (backendProcess) {
    console.log('停止后端服务');
    backendProcess.kill();
    backendProcess = null;
  }
}

// 应用准备就绪
app.whenReady().then(() => {
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

// 所有窗口关闭时退出应用
app.on('window-all-closed', () => {
  stopBackendService();
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

// 应用即将退出
app.on('before-quit', () => {
  stopBackendService();
});

// 处理证书错误
app.on('certificate-error', (event, webContents, url, error, certificate, callback) => {
  if (url.startsWith('http://localhost')) {
    // 忽略本地开发服务器的证书错误
    event.preventDefault();
    callback(true);
  } else {
    callback(false);
  }
});
