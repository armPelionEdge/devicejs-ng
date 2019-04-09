package routes_test
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
    "errors"
    "encoding/json"
    "net/http"
    "net/http/httptest"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    . "devicedb/routes"

    "github.com/gorilla/mux"
)

var _ = Describe("Snapshot", func() {
    var router *mux.Router
    var snapshotEndpoint *SnapshotEndpoint
    var clusterFacade *MockClusterFacade

    BeforeEach(func() {
        clusterFacade = &MockClusterFacade{ }
        router = mux.NewRouter()
        snapshotEndpoint = &SnapshotEndpoint{
            ClusterFacade: clusterFacade,
        }
        snapshotEndpoint.Attach(router)
    })

    Describe("/snapshot", func() {
        Describe("POST", func() {
            Context("When LocalSnapshot() returns an error", func() {
                It("Should respond with status code http.StatusInternalServerError", func() {
                    clusterFacade.defaultLocalSnapshotError = errors.New("Some error")

                    req, err := http.NewRequest("POST", "/snapshot", nil)

                    Expect(err).Should(BeNil())

                    rr := httptest.NewRecorder()
                    router.ServeHTTP(rr, req)

                    Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
                })
            })

            Context("When LocalSnapshot() does not return an error", func() {
                It("Should respond with status code http.StatusOK", func() {
                    req, err := http.NewRequest("POST", "/snapshot", nil)

                    Expect(err).Should(BeNil())

                    rr := httptest.NewRecorder()
                    router.ServeHTTP(rr, req)

                    Expect(rr.Code).Should(Equal(http.StatusOK))
                })

                It("Should respond with a body that is the JSON encoded snapshot returned by LocalSnapshot()", func() {
                    clusterFacade.defaultLocalSnapshotResponse = Snapshot{
                        UUID: "abc",
                        Status: SnapshotProcessing,
                    }

                    req, err := http.NewRequest("POST", "/snapshot", nil)

                    Expect(err).Should(BeNil())

                    rr := httptest.NewRecorder()
                    router.ServeHTTP(rr, req)

                    var snapshot Snapshot

                    Expect(json.Unmarshal(rr.Body.Bytes(), &snapshot)).Should(BeNil())
                    Expect(snapshot).Should(Equal(clusterFacade.defaultLocalSnapshotResponse))
                })
            })
        })
    })
})
