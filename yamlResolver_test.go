package yamlResolver

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("yamlResolver", t, func() {
		resolver := YamlResolver{}

		Convey("error on bad root file", func() {
			err := resolver.LoadFile("./doesNotExist.yaml")
			So(err, ShouldNotBeNil)
		})

		Convey("error on bad file reference", func() {
			err := resolver.LoadFile("./testFiles/fileWithBadReference.yaml")
			So(err, ShouldNotBeNil)
		})

		Convey("error on circular reference", func() {
			err := resolver.LoadFile("./testFiles/circular/parent.yaml")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "circular reference detected")
		})

		Convey("simple", func() {
			err := resolver.LoadFile("./testFiles/simple/parent.yaml")
			So(err, ShouldBeNil)
			So(resolver.String(), ShouldEqual, expectedSimple)
			err = resolver.SaveFile("test.yaml")
			fmt.Println(err)
		})

		Convey("simpleArray", func() {
			err := resolver.LoadFile("./testFiles/simpleArray/parent.yaml")
			So(err, ShouldBeNil)
			So(resolver.String(), ShouldEqual, expectedSimpleArray)
		})

		Convey("family", func() {
			err := resolver.LoadFile("./testFiles/family/gen1/grandNana.yaml")
			So(err, ShouldBeNil)
			So(resolver.String(), ShouldEqual, expectedFamily)
		})

		Convey("swagger", func() {
			err := resolver.LoadFile("./testFiles/swaggerSpec/index.yaml")
			So(err, ShouldBeNil)
			So(resolver.String(), ShouldEqual, expectedSwagger)
		})

	})
}

// Expected resolved test data below
var expectedSimple = clean(`
name: Nana
age: 75
child:
  name: Michael
  age: 58
`)

var expectedSimpleArray = clean(`
name: Nana
age: 75
children:
- name: Michael
  age: 58
- name: Bev
  age: 55
`)

var expectedFamily = clean(`
name: Grand Nana
age: 92
children:
- name: Nana
  age: 75
  husband:
    name: Pops
    age: 80
    children:
    - name: Michael
      age: 59
      children:
      - name: Chris
        age: 30
      - name: Erin
        age: 26
      - name: Sally
        age: 21
      - name: Elizabeth
        age: 21
    - name: Bev
      age: null
    - name: Carol
      age: null
  children:
  - name: Michael
    age: 59
    children:
    - name: Chris
      age: 30
    - name: Erin
      age: 26
    - name: Sally
      age: 21
    - name: Elizabeth
      age: 21
  - name: Bev
    age: null
  - name: Carol
    age: null
`)

var expectedSwagger = clean(`
swagger: '2.0'
##################################################################################################
#                               API INFORMATION                                                  #
##################################################################################################
info:
  title: Case Service API
  version: v1
  contact:
    url: http://github.schq.secious.com/WebUI/CaseService/wiki/Public-API-Docs
  description: Logrhythm Case Service API

##################################################################################################
#                         HOST/BASEPATH/CONTENT-TYPES                                            #
##################################################################################################
host: localhost:3000
# array of all schemes that your API supports
schemes:
  - https
# will be prefixed to all paths
basePath: /api
# perhaps this is the return type for responses? Makes sense for a simple API that just returns data in JSON form
produces:
  - application/json


##################################################################################################
#                                         SECURITY                                               #
##################################################################################################

securityDefinitions:
  Bearer:
    type: apiKey
    name: Authorization
    in: header

##################################################################################################
#                                         PATHS                                                  #
##################################################################################################

paths:

  #########################
  #     UPSERT A CASE    #
  ########################
  /cases:
    x-swagger-router-controller: case
    ### POST /cases ###
    post:
      operationId: upsertCase
      security:
        - Bearer: []
      description: |
        To update a specific case, you can provide id, caseNumber, or externalId in the body of the request.
        Those fields will be searched in that order to try and find an existing case. If no matches are found using id and caseNumber, the system returns a "404 not found" error.
        If no match is found when an externalId is the only identifier provided, a new case with that externalId will be created.
        If a match on id or number is found AND an externalId is specified, the request is treated as an update request where the externalId provided will replace the old one
      summary: upsert a case
      parameters:
        - name: body
          in: body
          description: the body of the case which includes the id of a case (if doing update) and either all the attributes (for a create) or some attributes (for an update)
          required: true
          schema:
            type: object
            required: [id, externalId]
            properties:
              priority:
                type: number
                description: priority of the case
                enum: [
                1,2,3,4,5
                ]
              summary:
                type: string
                description: summary of the case
                maxLength: 4000
              number:
                type: number
                description: case number displayed in the UI, generated by LogRhyhtm
              name:
                type: string
                description: name of the case
                maxLength: 250
              externalId:
                type: string
                description: custom defined unique identifier, can be thought of like a foreign key
                maxLength: 80
              id:
                type: string
                format: uuid
                description: case GUID generated by LogRhyhtm
      responses:
        201:
          description: The newly created case
          schema:
            title: 'Complete Case'
            $ref: '#/definitions/complete_case'
          # examples:
          #     $ref: 'responses/examples/complete_case.example.yaml'
        200:
          description: The upserted case
          schema:
            title: 'Complete Case'
            $ref: '#/definitions/complete_case'
          # examples:
          #     $ref: 'responses/examples/complete_case.example.yaml'



  #########################
  #      GET A CASE      #
  ########################
  /cases/{id}:
    x-swagger-router-controller: case
    ### GET /cases/{id} ###
    get:
      operationId: getCaseDetails
      security:
        - Bearer: []
      description: Get the details of a case by id
      summary: Get the details of a case by id
      parameters:
        - name: id
          in: path
          required: true
          type: string
          format: uuid
          description: case GUID generated by LogRhyhtm
      responses:
        200:
          description: The returned case
          schema:
            title: 'Complete Case'
            $ref: '#/definitions/complete_case'
          # examples:
          #     $ref: 'responses/examples/complete_case.example.yaml'



  ##################################
  #    CHANGE STATUS OF A CASE     #
  ##################################

  /cases/{id}/actions/changeStatus:
    x-swagger-router-controller: case
    ### PUT /cases/{id}/actions/changeStatus
    put:
      operationId: changeStatus
      security:
        - Bearer: []
      description: Change the state of a case referenced by id
      summary: Change the state of a case referenced by id
      parameters:
        - name: id
          in: path
          required: true
          type: string
          format: uuid
          description: case GUID generated by LogRhyhtm
        - name: body
          in: body
          description: A body that contains the new statusName to be given to the case with id id
          required: true
          schema:
            type: object
            required: [statusName]
            properties:
              statusName:
                type: string
                description: The name of the case status
                enum: [
                "Created",
                "Completed",
                "Incident",
                "Mitigated",
                "Resolved",
                ]
      responses:
        200:
          description: The updated case
          schema:
            title: 'Complete Case'
            $ref: '#/definitions/complete_case'
          # examples:
          #     $ref: 'responses/examples/complete_case.example.yaml'



  #######################################
  #       ADD EVIDENCE TO A CASE        #
  #######################################
  /cases/{id}/evidence/note:
    x-swagger-router-controller: case
    ### POST /cases/{id}/evidence/note
    post:
      operationId: createNote
      security:
        - Bearer: []
      description: add a note to a case referenced by id
      summary: add a note to a case referenced by id
      parameters:
        - name: id
          in: path
          required: true
          type: string
          format: uuid
          description: case GUID generated by LogRhyhtm
        - name: body
          in: body
          description: A body that contains the new statusName to be given to the case with id id
          required: true
          schema:
            type: object
            required: [text]
            properties:
              text:
                type: string
                description: The content of the evidence note
                maxLength: 250
      responses:
        200:
          description: An empty body is returned
          schema: {}

definitions:
  complete_case:
    required: [priority, summary, statusName, number, name, externalId, id]
    properties:
      priority:
        type: number
        description: priority of the case
        enum: [
        1,2,3,4,5
        ]
      summary:
        type: string
        description: summary of the case
        maxLength: 4000
      statusName:
        type: string
        description: The name of the case status
        enum: [
        "Created",
        "Completed",
        "Incident",
        "Mitigated",
        "Resolved",
        ]
      number:
        type: number
        description: case number displayed in the UI, generated by LogRhyhtm
      name:
        type: string
        description: name of the case
        maxLength: 250
      externalId:
        type: string
        description: custom defined unique identifier, can be thought of like a foreign key
        maxLength: 80
      id:
        type: string
        format: uuid
        description: case GUID generated by LogRhyhtm
`)

//just using this to make test data look cleaner
func clean(s string) string {
	return strings.Trim(s, "\n")
}
