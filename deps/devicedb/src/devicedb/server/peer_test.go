package server_test
//
 // Copyright (c) 2019 ARM Limited.
 //
 // SPDX-License-Identifier: MIT
 //
 // Permission is hereby granted, free of charge, to any person obtaining a copy
 // of this software and associated documentation files (the "Software"), to
 // deal in the Software without restriction, including without limitation the
 // rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 // sell copies of the Software, and to permit persons to whom the Software is
 // furnished to do so, subject to the following conditions:
 //
 // The above copyright notice and this permission notice shall be included in all
 // copies or substantial portions of the Software.
 //
 // THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 // IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 // FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 // AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 // LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 // OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 // SOFTWARE.
 //


import (
    "fmt"
    "time"
    "errors"
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    
    . "devicedb/server"
    . "devicedb/util"
    . "devicedb/bucket"
    . "devicedb/data"
    ddbSync "devicedb/sync"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

func loadCerts(id string) (*tls.Config, *tls.Config, error) {
    clientCertificate, err := tls.LoadX509KeyPair("../test_certs/" + id + ".client.cert.pem", "../test_certs/" + id + ".client.key.pem")
    
    if err != nil {
        return nil, nil, err
    }
    
    serverCertificate, err := tls.LoadX509KeyPair("../test_certs/" + id + ".server.cert.pem", "../test_certs/" + id + ".server.key.pem")
    
    if err != nil {
        return nil, nil, err
    }
    
    rootCAChain, err := ioutil.ReadFile("../test_certs/ca-chain.cert.pem")
    
    if err != nil {
        return nil, nil, err
    }
    
    rootCAs := x509.NewCertPool()
    if !rootCAs.AppendCertsFromPEM(rootCAChain) {
        return nil, nil, errors.New("Could not append certs to chain")
    }
    
    var serverTLSConfig = &tls.Config{
        Certificates: []tls.Certificate{ serverCertificate },
        ClientCAs: rootCAs,
    }
    var clientTLSConfig = &tls.Config{
        Certificates: []tls.Certificate{ clientCertificate },
        RootCAs: rootCAs,
    }
    
    return serverTLSConfig, clientTLSConfig, nil
}

const SYNC_PERIOD_MS = 10

var _ = Describe("Hub", func() {
    var initiatorHub *Hub
    var responderHub *Hub
    var neutralHub *Hub
    var initiatorSyncController *SyncController
    var responderSyncController *SyncController
    var neutralSyncController *SyncController
    var initiatorServer *Server
    var responderServer *Server
    var neutralServer *Server
    
    responderServerTLS, responderClientTLS, err := loadCerts("WWRL000000")

    if err != nil {
        fmt.Println("Unable to load responder certs", err)
        
        return
    }
    
    initiatorServerTLS, initiatorClientTLS, err := loadCerts("WWRL000001")
    
    if err != nil {
        fmt.Println("Unable to load initiator certs", err)
        
        return
    }
    
    stop := make(chan int)
    
    BeforeEach(func() {
        responderSyncController = NewSyncController(2, nil, ddbSync.NewPeriodicSyncScheduler(SYNC_PERIOD_MS), 1000)
        responderHub = NewHub("", responderSyncController, responderClientTLS)
        responderServer, _ = NewServer(ServerConfig{
            DBFile: "/tmp/testdb-" + RandomString(),
            Port: 8080,
            ServerTLS: responderServerTLS,
            Hub: responderHub,
        })
        
        initiatorSyncController = NewSyncController(2, nil, ddbSync.NewPeriodicSyncScheduler(SYNC_PERIOD_MS), 1000)
        initiatorHub = NewHub("", initiatorSyncController, initiatorClientTLS)
        initiatorServer, _ = NewServer(ServerConfig{
            DBFile: "/tmp/testdb-" + RandomString(),
            Port: 8181,
            ServerTLS: initiatorServerTLS,
            Hub: initiatorHub,
        })
        
        neutralSyncController = NewSyncController(2, nil, ddbSync.NewPeriodicSyncScheduler(SYNC_PERIOD_MS), 1000)
        neutralHub = NewHub("", neutralSyncController, initiatorClientTLS) // WWRL000001
        neutralServer, _ = NewServer(ServerConfig{
            DBFile: "/tmp/testdb-" + RandomString(),
            Port: 8282,
            ServerTLS: initiatorServerTLS,
            Hub: neutralHub,
        })
        
        go func() {
            responderServer.Start()
            stop <- 1
        }()
        
        time.Sleep(time.Millisecond * 100)
    })
    
    AfterEach(func() {
        responderServer.Stop()
        <-stop
    })
    
    Describe("sync", func() {
        It("makes sure that the id is extracted correctly from the client certificate and server certificates", func() {
            initiatorHub.Connect("WWRL000000", "127.0.0.1", 8080)
            //responderSyncController.StartResponderSessions()
            //initiatorSyncController.StartInitiatorSessions()
            responderSyncController.Start()
            initiatorSyncController.Start()
            
            go func() {
                for i := 0; i < 10; i += 1 {
                    time.Sleep(time.Second * 1)
                    updateBatch := NewUpdateBatch()
                    updateBatch.Put([]byte(RandomString()), []byte(RandomString()), NewDVV(NewDot("", 0), map[string]uint64{ }))
                    responderServer.Buckets().Get("default").Batch(updateBatch)
                }
            }()
            
            Expect(err).Should(BeNil())
            
            time.Sleep(time.Second * 60)
            
            initiatorHub.Disconnect("WWRL000000")
            
            time.Sleep(time.Second * 5)
            
            Expect(true).Should(BeTrue())
        })
        
        It("connect the same client twice.", func() {
            initiatorHub.Connect("WWRL000000", "127.0.0.1", 8080)
            neutralHub.Connect("WWRL000000", "127.0.0.1", 8080)
            
            time.Sleep(time.Second * 1)
            
            initiatorHub.Disconnect("WWRL000000")
            
            time.Sleep(time.Second * 1)
            
            Expect(true).Should(BeTrue())
        })
    })
})