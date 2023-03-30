/*
 *
 *     @file: Proxy.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/30 下午11:29
 *     @last modified: 2023/3/30 下午3:37
 *
 *
 *
 */

package Net

import (
	"GodDns/Util"
	"net/url"
)

type proxy = string

type Proxys []proxy

var GlobalProxys = &Proxys{}

func IsProxyValid(proxy proxy) bool {
	_, err := url.Parse(proxy)
	if err != nil {
		return false
	}
	return true
}

func AddProxy(target *Proxys, proxy ...proxy) {
	*target = append(*target, proxy...)
}

func AddProxy2Top(target *Proxys, proxy ...proxy) {
	*target = append(proxy, *target...)
}

func (p *Proxys) GetProxyIter() *Util.Iter[proxy] {
	return Util.NewIter[proxy]((*[]proxy)(p))
}
