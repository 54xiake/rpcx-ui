package main

import (
	"encoding/base64"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/54xiake/libkv"
	kvstore "github.com/54xiake/libkv/store"
	"github.com/54xiake/libkv/store/redis"
)

type RedisRegistry struct {
	kv kvstore.Store
}

func (r *RedisRegistry) initRegistry() {
	redis.Register()

	if strings.HasPrefix(serverConfig.ServiceBaseURL, "/") {
		serverConfig.ServiceBaseURL = serverConfig.ServiceBaseURL[1:]
	}

	if strings.HasSuffix(serverConfig.ServiceBaseURL, "/") {
		serverConfig.ServiceBaseURL = serverConfig.ServiceBaseURL[0 : len(serverConfig.ServiceBaseURL)-1]
	}

	kv, err := libkv.NewStore(kvstore.REDIS, []string{serverConfig.RegistryURL}, nil)
	if err != nil {
		log.Printf("cannot create redis registry: %v", err)
		return
	}
	r.kv = kv

	return
}

func (r *RedisRegistry) fetchServices() []*Service {
	var services []*Service

	//kvs, err := r.kv.List(serverConfig.ServiceBaseURL, &kvstore.ReadOptions{})
	kvs, err := r.kv.List("/"+serverConfig.ServiceBaseURL, &kvstore.ReadOptions{})
	if err != nil {
		log.Printf("failed to list services %s: %v", serverConfig.ServiceBaseURL, err)
		return services
	}

	for _, value := range kvs {
		serviceName := value.Key
		log.Printf("%#v", value.Key)

		//nodes, err := r.kv.List(serverConfig.ServiceBaseURL + "/" + value.Key, &kvstore.ReadOptions{})
		nodes, err := r.kv.List(value.Key, &kvstore.ReadOptions{})
		if err != nil {
			log.Printf("failed to list  %s: %v", serverConfig.ServiceBaseURL+"/"+value.Key, err)
			continue
		}

		for _, n := range nodes {
			var serviceAddr = n.Key

			v, err := url.ParseQuery(string(n.Value[:]))
			if err != nil {
				log.Println("redis value parse failed. error: ", err.Error())
				continue
			}
			state := "n/a"
			group := ""
			if err == nil {
				state = v.Get("state")
				if state == "" {
					state = "active"
				}
				group = v.Get("group")
			}
			id := base64.StdEncoding.EncodeToString([]byte(serviceName + "@" + serviceAddr))
			service := &Service{ID: id, Name: serviceName, Address: serviceAddr, Metadata: string(n.Value[:]), State: state, Group: group}
			services = append(services, service)
		}

	}

	return services
}

func (r *RedisRegistry) deactivateService(name, address string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)

	kv, err := r.kv.Get(key, &kvstore.ReadOptions{})

	if err != nil {
		return err
	}

	v, err := url.ParseQuery(string(kv.Value[:]))
	if err != nil {
		log.Println("redis value parse failed. err ", err.Error())
		return err
	}
	v.Set("state", "inactive")
	err = r.kv.Put(kv.Key, []byte(v.Encode()), &kvstore.WriteOptions{IsDir: false})
	if err != nil {
		log.Println("redis set failed, err : ", err.Error())
	}

	return err
}

func (r *RedisRegistry) activateService(name, address string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)
	kv, err := r.kv.Get(key, &kvstore.ReadOptions{})

	v, err := url.ParseQuery(string(kv.Value[:]))
	if err != nil {
		log.Println("redis value parse failed. err ", err.Error())
		return err
	}
	v.Set("state", "active")
	err = r.kv.Put(kv.Key, []byte(v.Encode()), &kvstore.WriteOptions{IsDir: false})
	if err != nil {
		log.Println("redis put failed. err: ", err.Error())
	}

	return err
}

func (r *RedisRegistry) updateMetadata(name, address string, metadata string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)
	err := r.kv.Put(key, []byte(metadata), &kvstore.WriteOptions{IsDir: false})
	return err
}
