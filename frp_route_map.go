// server
var rootCmd = &cobra.Command
=> runServer(cfg)
	func runServer(cfg config.ServerCommonConf) (err error)
	=> svr, err := server.NewService(cfg)
		=> ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.BindPort)) //服务开始监听指定端口
		=> svr.listener = ln // 保存服务的 listener 后面的函数中调用Accept 处理逻辑
	=> svr.Run()
		func (svr *Service) Run()
		=> svr.HandleListener(svr.listener)
			func (svr *Service) HandleListener(l net.Listener)
			=> for {c, err := l.Accept()} // 循环处理Accept
			=> go func(ctx context.Context, frpConn net.Conn) // 开启协程处理新连接请求， 之后的数据读写通过 frpConn 连接实例
				=> go svr.handleConnection(ctx, stream) 
					func (svr *Service) handleConnection(ctx context.Context, conn net.Conn) // 处理新连接的逻辑, conn = frpConn
					=> case *msg.Login // 接受frpc 调用login  rawMsg:&{0.33.0  linux amd64  ea304d5c057e5d3cb05aef21f3576baf 1593689899 18:c0:4d:28:b1:44 map[] 1}
					=> svr.RegisterControl(conn, m)
						func (svr *Service) RegisterControl(ctlConn net.Conn, loginMsg *msg.Login) (err error) // 实例化一个 Control , ctlConn = conn = frpConn
						=> svr.ctlManager.Add(loginMsg.RunId, ctl) //保存 实例化的 Control 
						=> ctl.Start()
							func (ctl *Control) Start() 
							=> for i := 0; i < ctl.poolCount; i++ {
									ctl.sendCh <- &msg.ReqWorkConn{}
								}  // 让客户端用新连接请求frps，打开新的workConn 开启三个协程
							=> go ctl.reader() // 读取数据协程, 将读取的数据写到 channel ctl.readCh
							=> go ctl.stoper() // 停止 Control 释放资源
							=> go ctl.manager() // 开启协程
								func (ctl *Control) manager()
								=> case rawMsg, ok := <-ctl.readCh // 从go ctl.reader()中的 readCh 读取数据
									=> case *msg.NewProxy  // 实例化一个新的pxy
									=> ctl.RegisterProxy(m)
										func (ctl *Control) RegisterProxy(pxyMsg *msg.NewProxy) (remoteAddr string, err error)
										=> pxy, err := proxy.NewProxy(ctl.ctx, userInfo, ctl.rc, ctl.poolCount, ctl.GetWorkConn, pxyConf, ctl.serverCfg) // 创建代理实例，传入之后需要调用的函数：ctl.GetWorkConn 获取连接
											func (ctl *Control) GetWorkConn() (workConn net.Conn, err error)
											=> case workConn, ok = <-ctl.workConnCh
										=> remoteAddr, err = pxy.Run()
										=> err = ctl.pxyManager.Add(pxyMsg.ProxyName, pxy) // 将 pxy加入到 ProxyManager.pxys 中
											func (pxy *TcpProxy) Run() (remoteAddr string, err error) // 根据收到的信息中的端口信息，启动新的监听端口
											=> listener, errRet := net.Listen("tcp", fmt.Sprintf("%s:%d", pxy.serverCfg.ProxyBindAddr, pxy.realPort)) // 开始监听
											=> pxy.listeners = append(pxy.listeners, listener)   // 保存 listener
											=> pxy.startListenHandler(pxy, HandleUserTcpConnection) 
												func (pxy *BaseProxy) startListenHandler(p Proxy, handler func(Proxy, net.Conn, config.ServerCommonConf))
												=> for _, listener := range pxy.listeners  //遍历 pxy.listeners
												=> go func(l net.Listener) {c, err := l.Accept()} // 给每个 listener 启动处理协程 Accept , Accept 内部 conn, ok := <-l.acceptCh 阻塞在channel 等待新连接到来
													=> go handler(p, c, pxy.serverCfg) // 这个 handler 就是 HandleUserTcpConnection
									=> case *msg.Ping // 心跳信息

					=> case *msg.NewWorkConn // frpc登录完成，然后发送此命令请求创建workConn， rawMsg:&{18:c0:4d:28:b1:44  0}
					svr.RegisterWorkConn(conn, m)
						func (svr *Service) RegisterWorkConn(workConn net.Conn, newMsg *msg.NewWorkConn) error
						=> ctl, exist := svr.ctlManager.GetById(newMsg.RunId) // 通过Runid获取已经注册的ctl
						=> ctl.RegisterWorkConn(workConn)
							func (ctl *Control) RegisterWorkConn(conn net.Conn) error
							=> case ctl.workConnCh <- conn  //将连接句柄 conn 写入到 连接buffer channel



				func HandleUserTcpConnection(pxy Proxy, userConn net.Conn, serverCfg config.ServerCommonConf)
				=> workConn, err := pxy.GetWorkConnFromPool(userConn.RemoteAddr(), userConn.LocalAddr()) // go handler 其实调用到这里
					func (pxy *BaseProxy) GetWorkConnFromPool(src, dst net.Addr) (workConn net.Conn, err error)
					=> if workConn, err = pxy.getWorkConnFn(); err != nil // 调用 ctl.GetWorkConn 获取连接
						func (ctl *Control) GetWorkConn() (workConn net.Conn, err error)
						=> case workConn, ok = <-ctl.workConnCh
				=> var local io.ReadWriteCloser = workConn
				=> inCount, outCount := frpIo.Join(local, userConn) // 开始做转发


ProxyManager：
func runServer(cfg config.ServerCommonConf) (err error)
=> svr, err := server.NewService(cfg)
	svr = &Service{pxyManager:    proxy.NewProxyManager()}
			func (svr *Service) RegisterControl(ctlConn net.Conn, loginMsg *msg.Login) (err error)
				ctl := NewControl(ctx, svr.rc, svr.pxyManager, svr.pluginManager, svr.authVerifier, ctlConn, loginMsg, svr.cfg) 
				=> return &Control{	pxyManager:      pxyManager}  //此时 ctl中的 pxyManager 使用的是全局的 pxyManager


// 思路：
// 1. 收到 frpc 连接，得到 frpConn
// 2. 接受处理 frpc login请求
// 3. 实例化 Control , 启动三个协程
// 4. 协程 manager 处理frpc发来的数据， 包括msg.NewProxy msg.Ping(心跳)



// client:
var rootCmd = &cobra.Command
=> err := runClient(cfgFile)
	func runClient(cfgFilePath string) (err error)
	=> startService(cfg, pxyCfgs, visitorCfgs, cfgFilePath)
		func startService(cfg config.ClientCommonConf, pxyCfgs map[string]config.ProxyConf, visitorCfgs map[string]config.VisitorConf, cfgFile string) (err error)
		=> svr, errRet := client.NewService(cfg, pxyCfgs, visitorCfgs, cfgFile)
		=> err = svr.Run()
			func (svr *Service) Run() error
			=> for{conn, session, err := svr.login();
				func (svr *Service) login() (conn net.Conn, session *fmux.Session, err error)
				=> conn, err = frpNet.ConnectServerByProxyWithTLS(svr.cfg.HttpProxy, svr.cfg.Protocol, fmt.Sprintf("%s:%d", svr.cfg.ServerAddr, svr.cfg.ServerPort), tlsConfig)
				=> session, err = fmux.Client(conn, fmuxCfg)
				=> 	stream, errRet := session.OpenStream(); conn = stream // 使用session的方式连接
				=> if err = msg.WriteMsg(conn, loginMsg) // 发送登录信息
				=> if err = msg.ReadMsgInto(conn, &loginRespMsg) // 接受登录返回信息
			=> ctl := NewControl(svr.ctx, svr.cfg.RunId, conn, session, svr.cfg, svr.pxyCfgs, svr.visitorCfgs, svr.serverUDPPort, svr.authSetter) // 参数中的session 即为login初始化过的session,, conn为session的一个stream
			=> ctl.Run()
				func (ctl *Control) Run()
				=> go ctl.worker() // 启动工作协程
					func (ctl *Control) worker()
					=> go ctl.msgHandler()
						func (ctl *Control) msgHandler()
						=> case <-hbSend.C // 发送心跳
						=> case rawMsg, ok := <-ctl.readCh 
							=> case *msg.ReqWorkConn // 新建workConn,请求frps
								=> go ctl.HandleReqWorkConn(m)
									func (ctl *Control) HandleReqWorkConn(inMsg *msg.ReqWorkConn) {
									=> workConn, err := ctl.connectServer() // 根据session创建新的连接 workConn
									=> if err = msg.WriteMsg(workConn, m); err != nil // 发送 m := &msg.NewWorkConn{RunId: ctl.runId}
									=> if err = msg.ReadMsgInto(workConn, &startMsg); err != nil // 此时已知阻塞在这里，直到有客户端连接到frps，然后frps 在 GetWorkConnFromPool 拿到对应的workConn，
													//然后发送 err := msg.WriteMsg(workConn, &msg.StartWorkConn{
																					// 	ProxyName: pxy.GetName(),
																					// 	SrcAddr:   srcAddr,
																					// 	SrcPort:   uint16(srcPort),
																					// 	DstAddr:   dstAddr,
																					// 	DstPort:   uint16(dstPort),
																					// 	Error:     "",
																					// })
													//收到frps的这条消息后，继续往下执行
									=> ctl.pm.HandleWorkConn(startMsg.ProxyName, workConn, &startMsg)
										func (pm *ProxyManager) HandleWorkConn(name string, workConn net.Conn, m *msg.StartWorkConn)
										=> pw, ok := pm.proxies[name]
										=> pw.InWorkConn(workConn, m)
											func (pw *ProxyWrapper) InWorkConn(workConn net.Conn, m *msg.StartWorkConn)
											=> go pxy.InWorkConn(workConn, m) // interface
												func (pxy *TcpMuxProxy) InWorkConn(conn net.Conn, m *msg.StartWorkConn)
												=> HandleTcpWorkConnection(pxy.ctx, &pxy.cfg.LocalSvrConf, pxy.proxyPlugin, &pxy.cfg.BaseProxyConf, pxy.limiter,conn, []byte(pxy.clientCfg.Token), m)
													func HandleTcpWorkConnection(ctx context.Context, localInfo *config.LocalSvrConf, proxyPlugin plugin.Plugin, baseInfo *config.BaseProxyConf, limiter *rate.Limiter, workConn net.Conn, encKey []byte, m *msg.StartWorkConn)
													=> localConn, err := frpNet.ConnectServer("tcp", fmt.Sprintf("%s:%d", localInfo.LocalIp, localInfo.LocalPort))
													=> frpIo.Join(localConn, remote) // 做数据转发
							=> case *msg.NewProxyResp
								=> ctl.HandleNewProxyResp(m)
							=> case *msg.Pong //处理心跳
					=> go ctl.reader()
					=> go ctl.writer()
				=> ctl.pm.Reload(ctl.pxyCfgs) // start all proxies
					func (pm *ProxyManager) Reload(pxyCfgs map[string]config.ProxyConf)
					=> for name, cfg := range pxyCfgs {
							if _, ok := pm.proxies[name]; !ok {
								pxy := NewProxyWrapper(pm.ctx, cfg, pm.clientCfg, pm.HandleEvent, pm.serverUDPPort)
								pm.proxies[name] = pxy
								addPxyNames = append(addPxyNames, name)

								pxy.Start()
							}
						}		



// PoolCount specifies the number of connections the client will make to
// the server in advance. By default, this value is 0.
	PoolCount int `json:"pool_count"`






type TcpProxy struct {
	*BaseProxy
	cfg *config.TcpProxyConf

	realPort int
}



func NewProxy(ctx context.Context, userInfo plugin.UserInfo, rc *controller.ResourceController, poolCount int, getWorkConnFn GetWorkConnFn, pxyConf config.ProxyConf, serverCfg config.ServerCommonConf) (pxy Proxy, err error)
=> switch cfg := pxyConf.(type)
	case *config.TcpProxyConf