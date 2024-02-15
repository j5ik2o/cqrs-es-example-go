// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Junichi Kato"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/": {
            "get": {
                "description": "Index",
                "summary": "Index",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/add-member": {
            "post": {
                "description": "Add a member to a group chat",
                "summary": "Add a member to a group chat",
                "parameters": [
                    {
                        "description": "AddMemberRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.AddMemberRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.AddMemberResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/create": {
            "post": {
                "description": "Create a group chat",
                "summary": "Create a group chat",
                "parameters": [
                    {
                        "description": "CreateGroupChatRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.CreateGroupChatRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.CreateGroupChatResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/delete": {
            "post": {
                "description": "Delete a group chat",
                "summary": "Delete a group chat",
                "parameters": [
                    {
                        "description": "DeleteGroupChatRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.DeleteGroupChatRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.DeleteGroupChatResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/delete-message": {
            "post": {
                "description": "Delete a message from a group chat",
                "summary": "Delete a message from a group chat",
                "parameters": [
                    {
                        "description": "DeleteMessageRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.DeleteMessageRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.DeleteMessageResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/post-message": {
            "post": {
                "description": "Post a message to a group chat",
                "summary": "Post a message to a group chat",
                "parameters": [
                    {
                        "description": "PostMessageRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.PostMessageRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.PostMessageResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/remove-member": {
            "post": {
                "description": "Remove a member from a group chat",
                "summary": "Remove a member from a group chat",
                "parameters": [
                    {
                        "description": "RemoveMemberRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.RemoveMemberRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.RemoveMemberResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        },
        "/v1/group-chat/rename": {
            "post": {
                "description": "Rename a group chat",
                "summary": "Rename a group chat",
                "parameters": [
                    {
                        "description": "RenameGroupChatRequestBody",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.RenameGroupChatRequestBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.RenameGroupChatResponseSuccessBody"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api.GroupChatResponseErrorBody"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.AddMemberRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string"
                },
                "group_chat_id": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                },
                "user_account_id": {
                    "type": "string"
                }
            }
        },
        "api.AddMemberResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string"
                }
            }
        },
        "api.CreateGroupChatRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string",
                    "example": "UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z"
                },
                "name": {
                    "type": "string",
                    "example": "group-chat-name-1"
                }
            }
        },
        "api.CreateGroupChatResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string",
                    "example": "GroupChat-01H42K4ABWQ5V2XQEP3A48VE0Z"
                }
            }
        },
        "api.DeleteGroupChatRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string"
                },
                "group_chat_id": {
                    "type": "string"
                }
            }
        },
        "api.DeleteGroupChatResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string"
                }
            }
        },
        "api.DeleteMessageRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string"
                },
                "group_chat_id": {
                    "type": "string"
                },
                "message_id": {
                    "type": "string"
                },
                "user_account_id": {
                    "type": "string"
                }
            }
        },
        "api.DeleteMessageResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string"
                }
            }
        },
        "api.GroupChatResponseErrorBody": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "api.PostMessageRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string"
                },
                "group_chat_id": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "user_account_id": {
                    "type": "string"
                }
            }
        },
        "api.PostMessageResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string"
                },
                "message_id": {
                    "type": "string"
                }
            }
        },
        "api.RemoveMemberRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string"
                },
                "group_chat_id": {
                    "type": "string"
                },
                "user_account_id": {
                    "type": "string"
                }
            }
        },
        "api.RemoveMemberResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string"
                }
            }
        },
        "api.RenameGroupChatRequestBody": {
            "type": "object",
            "properties": {
                "executor_id": {
                    "type": "string"
                },
                "group_chat_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "api.RenameGroupChatResponseSuccessBody": {
            "type": "object",
            "properties": {
                "group_chat_id": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "GroupChat Write API",
	Description:      "This is Write API Server for GroupChat.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}