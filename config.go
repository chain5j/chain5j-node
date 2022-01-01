// Package node
//
// @author: xwc1125
package node

var (
	ConfPathName     = "conf"
	DataPathName     = "chaindata"
	P2pPathName      = "p2p"
	KeystorePathName = "keystore"
	CertsPathName    = "certs"
	IPCPathName      = "chain5j.ipc"

	// sn配置
	isSNVerify = true
	snPubKey   = ""
)

//type ChainDirConfig struct {
//	RootDir     string // 根目录
//	ConfDir     string // 配置目录
//	DBDir       string // 数据目录
//	P2PDir      string // 数据目录
//	KeystoreDir string // 密码目录
//	IPCPath     string // IPC path
//	CertsDir    string // 证书目录
//}
//
//func GetDataDir() *ChainDirConfig {
//	if chainDirConfig != nil {
//		return chainDirConfig
//	}
//	var dataDir = "./"
//	//if cli.RootCli != nil {
//	//	dataDir = cli.RootCli.DataDir
//	//}
//	return SetDataDir(dataDir)
//}
//
//func SetDataDir(rootDir string) *ChainDirConfig {
//	rootDir = strings.TrimSuffix(rootDir, "/")
//	confPath := filepath.Join(rootDir, ConfPathName)
//	dataPath := filepath.Join(rootDir, DataPathName)
//	p2pPath := filepath.Join(rootDir, P2pPathName)
//	keystorePath := filepath.Join(rootDir, KeystorePathName)
//	certsPath := filepath.Join(rootDir, CertsPathName)
//	ipcPath := filepath.Join(rootDir, IPCPathName)
//
//	if !ioutil.PathExists(confPath) {
//		ioutil.MakeDirAll(confPath)
//	}
//	if !ioutil.PathExists(dataPath) {
//		ioutil.MakeDirAll(dataPath)
//	}
//	if !ioutil.PathExists(p2pPath) {
//		ioutil.MakeDirAll(p2pPath)
//	}
//	if !ioutil.PathExists(certsPath) {
//		ioutil.MakeDirAll(certsPath)
//	}
//	if !ioutil.PathExists(keystorePath) {
//		ioutil.MakeDirAll(keystorePath)
//	}
//	chainDirConfig = &ChainDirConfig{
//		RootDir:     rootDir,
//		ConfDir:     confPath,
//		DBDir:       dataPath,
//		P2PDir:      p2pPath,
//		CertsDir:    certsPath,
//		KeystoreDir: keystorePath,
//		IPCPath:     ipcPath,
//	}
//
//	logger.Info("chainDirConfig", "path", chainDirConfig)
//	return chainDirConfig
//}

//// 节点配置
//type NodeConfig struct {
//	IpcConfig    rpc.IpcConfig     `json:"ipc" mapstructure:"ipc"`                     // ipc
//	TlsConfig    network.TlsConfig `json:"tls" mapstructure:"tls"`                     // tls
//	HttpConfig   rpc.HttpConfig    `json:"http" mapstructure:"http"`                   // http
//	WSConfig     rpc.WSConfig      `json:"ws" mapstructure:"ws"`                       // websocket
//	P2P          models.P2PConfig  `json:"p2p" mapstructure:"p2p"`                     // p2p
//	DBPath       string            `json:"db_path" mapstructure:"db_path"`             // db 路径
//	EnableWorker bool              `json:"enable_worker" mapstructure:"enable_worker"` // 是否启动worker
//}
//
//func GetNodeConfig() (*NodeConfig, error) {
//	var err error
//	// ipc
//	{
//		ipcName := viper.Sub("ipc").GetString("ipc_name")
//		if ipcName != "" && strings.HasSuffix(ipcName, ".ipc") {
//			IPCPathName = ipcName
//		}
//	}
//	// tls
//	tlsConfig := network.TlsConfig{}
//	{
//		err = viper.UnmarshalKey("tls", &tlsConfig)
//		if err != nil {
//			logger.Error("viper.UnmarshalKey(tls, &tlsConfig) err", "err", err)
//		}
//		if tlsConfig.KeyFile != "" && !filepath.IsAbs(tlsConfig.KeyFile) {
//			tlsConfig.KeyFile = filepath.Join(GetDataDir().RootDir, tlsConfig.KeyFile)
//		}
//		if tlsConfig.CertFile != "" && !filepath.IsAbs(tlsConfig.CertFile) {
//			tlsConfig.CertFile = filepath.Join(GetDataDir().RootDir, tlsConfig.CertFile)
//		}
//		for i, root := range tlsConfig.CaRoots {
//			if root != "" && !filepath.IsAbs(root) {
//				tlsConfig.CaRoots[i] = filepath.Join(GetDataDir().RootDir, root)
//			}
//		}
//	}
//	// http
//	httpConfig := rpc.HttpConfig{}
//	{
//		err = viper.UnmarshalKey("http", &httpConfig)
//		if err != nil {
//			logger.Error("viper.UnmarshalKey(http, &httpConfig) err", "err", err)
//		}
//
//		// 解析cmd中的参数
//		host := viper.GetString("rpchost")
//		if host == "" {
//			httpConfig.Host = host
//		}
//		if httpConfig.Host == "" {
//			httpConfig.Host = "0.0.0.0"
//		}
//		port := viper.GetInt("rpcport")
//		if port != 9545 {
//			httpConfig.Port = port
//		}
//		if httpConfig.Port == 0 {
//			httpConfig.Port = port
//		}
//	}
//	// ws
//	wsConfig := rpc.WSConfig{}
//	{
//		err = viper.UnmarshalKey("ws", &wsConfig)
//		if err != nil {
//			logger.Error("viper.UnmarshalKey(ws, &wsConfig) err", "err", err)
//		}
//	}
//	// p2p2
//	p2pConfig := models.P2PConfig{}
//	{
//		err = viper.UnmarshalKey("p2p", &p2pConfig)
//		if err != nil {
//			logger.Error("viper.UnmarshalKey(p2p, &p2pConfig) err", "err", err)
//		}
//		if p2pConfig.MaxPeers == 0 {
//			p2pConfig.MaxPeers = p2p.MaxPeers
//		}
//		if !filepath.IsAbs(p2pConfig.KeyPath) {
//			p2pConfig.KeyPath = filepath.Join(GetDataDir().RootDir, p2pConfig.KeyPath)
//		}
//		if !filepath.IsAbs(p2pConfig.CertPath) {
//			p2pConfig.CertPath = filepath.Join(GetDataDir().RootDir, p2pConfig.CertPath)
//		}
//		if p2pConfig.CaRoots != nil && len(p2pConfig.CaRoots) > 0 {
//			for i, root := range p2pConfig.CaRoots {
//				if !filepath.IsAbs(root) {
//					root = filepath.Join(GetDataDir().RootDir, root)
//				}
//				caRootBytes, err := nioutil.ReadFile(root)
//				if err != nil {
//					return nil, err
//				}
//				p2pConfig.CaRoots[i] = string(caRootBytes)
//			}
//		}
//	}
//
//	return &NodeConfig{
//		IpcConfig: rpc.IpcConfig{
//			Path: GetDataDir().IPCPath,
//		},
//		TlsConfig:    tlsConfig,
//		HttpConfig:   httpConfig,
//		WSConfig:     wsConfig,
//		P2P:          p2pConfig,
//		DBPath:       GetDataDir().DBDir,
//		EnableWorker: viper.GetBool("mine"),
//	}, nil
//}
//
//func RegisterSNPubkey(pubKey string) {
//	snPubKey = pubKey
//}

//func GetLicense() (*license2.SN, error) {
//	input := viper.GetString("license")
//	if len(input) == 0 {
//		return &license2.SN{}, nil
//	}
//
//	//logger.Info("load license", "input", input)
//	jsonBytes, err := base64.StdEncoding.DecodeString(input)
//	if err != nil {
//		return nil, err
//	}
//	logger.Info("load license", "license", string(jsonBytes))
//
//	var sn license2.SN
//	err = json.Unmarshal(jsonBytes, &sn)
//	if err != nil {
//		return nil, errors.New("invalid license")
//	}
//
//	if !license2.Verify(sn.GetSignJson(), sn.Signature, []byte(snPubKey)) {
//		return nil, errors.New("invalid license")
//	}
//
//	return &sn, nil
//}
