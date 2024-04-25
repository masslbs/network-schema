/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots.tutorial || ($protobuf.roots.tutorial = {});

export const tutorial = $root.tutorial = (() => {

    /**
     * Namespace tutorial.
     * @exports tutorial
     * @namespace
     */
    const tutorial = {};

    tutorial.Person = (function() {

        /**
         * Properties of a Person.
         * @memberof tutorial
         * @interface IPerson
         * @property {string|null} [name] Person name
         * @property {number|null} [id] Person id
         * @property {number|null} [age] Person age
         */

        /**
         * Constructs a new Person.
         * @memberof tutorial
         * @classdesc Represents a Person.
         * @implements IPerson
         * @constructor
         * @param {tutorial.IPerson=} [properties] Properties to set
         */
        function Person(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Person name.
         * @member {string} name
         * @memberof tutorial.Person
         * @instance
         */
        Person.prototype.name = "";

        /**
         * Person id.
         * @member {number} id
         * @memberof tutorial.Person
         * @instance
         */
        Person.prototype.id = 0;

        /**
         * Person age.
         * @member {number} age
         * @memberof tutorial.Person
         * @instance
         */
        Person.prototype.age = 0;

        /**
         * Creates a new Person instance using the specified properties.
         * @function create
         * @memberof tutorial.Person
         * @static
         * @param {tutorial.IPerson=} [properties] Properties to set
         * @returns {tutorial.Person} Person instance
         */
        Person.create = function create(properties) {
            return new Person(properties);
        };

        /**
         * Encodes the specified Person message. Does not implicitly {@link tutorial.Person.verify|verify} messages.
         * @function encode
         * @memberof tutorial.Person
         * @static
         * @param {tutorial.IPerson} message Person message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Person.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.id);
            if (message.age != null && Object.hasOwnProperty.call(message, "age"))
                writer.uint32(/* id 3, wireType 0 =*/24).int32(message.age);
            return writer;
        };

        /**
         * Encodes the specified Person message, length delimited. Does not implicitly {@link tutorial.Person.verify|verify} messages.
         * @function encodeDelimited
         * @memberof tutorial.Person
         * @static
         * @param {tutorial.IPerson} message Person message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Person.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Person message from the specified reader or buffer.
         * @function decode
         * @memberof tutorial.Person
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {tutorial.Person} Person
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Person.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.tutorial.Person();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.name = reader.string();
                        break;
                    }
                case 2: {
                        message.id = reader.int32();
                        break;
                    }
                case 3: {
                        message.age = reader.int32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Person message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof tutorial.Person
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {tutorial.Person} Person
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Person.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Person message.
         * @function verify
         * @memberof tutorial.Person
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Person.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isInteger(message.id))
                    return "id: integer expected";
            if (message.age != null && message.hasOwnProperty("age"))
                if (!$util.isInteger(message.age))
                    return "age: integer expected";
            return null;
        };

        /**
         * Creates a Person message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof tutorial.Person
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {tutorial.Person} Person
         */
        Person.fromObject = function fromObject(object) {
            if (object instanceof $root.tutorial.Person)
                return object;
            let message = new $root.tutorial.Person();
            if (object.name != null)
                message.name = String(object.name);
            if (object.id != null)
                message.id = object.id | 0;
            if (object.age != null)
                message.age = object.age | 0;
            return message;
        };

        /**
         * Creates a plain object from a Person message. Also converts values to other types if specified.
         * @function toObject
         * @memberof tutorial.Person
         * @static
         * @param {tutorial.Person} message Person
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Person.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.name = "";
                object.id = 0;
                object.age = 0;
            }
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.age != null && message.hasOwnProperty("age"))
                object.age = message.age;
            return object;
        };

        /**
         * Converts this Person to JSON.
         * @function toJSON
         * @memberof tutorial.Person
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Person.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Person
         * @function getTypeUrl
         * @memberof tutorial.Person
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Person.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/tutorial.Person";
        };

        return Person;
    })();

    tutorial.Robot = (function() {

        /**
         * Properties of a Robot.
         * @memberof tutorial
         * @interface IRobot
         * @property {string|null} [name] Robot name
         * @property {number|null} [id] Robot id
         * @property {Array.<tutorial.Robot.IFeature>|null} [features] Robot features
         */

        /**
         * Constructs a new Robot.
         * @memberof tutorial
         * @classdesc Represents a Robot.
         * @implements IRobot
         * @constructor
         * @param {tutorial.IRobot=} [properties] Properties to set
         */
        function Robot(properties) {
            this.features = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Robot name.
         * @member {string} name
         * @memberof tutorial.Robot
         * @instance
         */
        Robot.prototype.name = "";

        /**
         * Robot id.
         * @member {number} id
         * @memberof tutorial.Robot
         * @instance
         */
        Robot.prototype.id = 0;

        /**
         * Robot features.
         * @member {Array.<tutorial.Robot.IFeature>} features
         * @memberof tutorial.Robot
         * @instance
         */
        Robot.prototype.features = $util.emptyArray;

        /**
         * Creates a new Robot instance using the specified properties.
         * @function create
         * @memberof tutorial.Robot
         * @static
         * @param {tutorial.IRobot=} [properties] Properties to set
         * @returns {tutorial.Robot} Robot instance
         */
        Robot.create = function create(properties) {
            return new Robot(properties);
        };

        /**
         * Encodes the specified Robot message. Does not implicitly {@link tutorial.Robot.verify|verify} messages.
         * @function encode
         * @memberof tutorial.Robot
         * @static
         * @param {tutorial.IRobot} message Robot message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Robot.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.id);
            if (message.features != null && message.features.length)
                for (let i = 0; i < message.features.length; ++i)
                    $root.tutorial.Robot.Feature.encode(message.features[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified Robot message, length delimited. Does not implicitly {@link tutorial.Robot.verify|verify} messages.
         * @function encodeDelimited
         * @memberof tutorial.Robot
         * @static
         * @param {tutorial.IRobot} message Robot message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Robot.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Robot message from the specified reader or buffer.
         * @function decode
         * @memberof tutorial.Robot
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {tutorial.Robot} Robot
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Robot.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.tutorial.Robot();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.name = reader.string();
                        break;
                    }
                case 2: {
                        message.id = reader.int32();
                        break;
                    }
                case 3: {
                        if (!(message.features && message.features.length))
                            message.features = [];
                        message.features.push($root.tutorial.Robot.Feature.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Robot message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof tutorial.Robot
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {tutorial.Robot} Robot
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Robot.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Robot message.
         * @function verify
         * @memberof tutorial.Robot
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Robot.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isInteger(message.id))
                    return "id: integer expected";
            if (message.features != null && message.hasOwnProperty("features")) {
                if (!Array.isArray(message.features))
                    return "features: array expected";
                for (let i = 0; i < message.features.length; ++i) {
                    let error = $root.tutorial.Robot.Feature.verify(message.features[i]);
                    if (error)
                        return "features." + error;
                }
            }
            return null;
        };

        /**
         * Creates a Robot message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof tutorial.Robot
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {tutorial.Robot} Robot
         */
        Robot.fromObject = function fromObject(object) {
            if (object instanceof $root.tutorial.Robot)
                return object;
            let message = new $root.tutorial.Robot();
            if (object.name != null)
                message.name = String(object.name);
            if (object.id != null)
                message.id = object.id | 0;
            if (object.features) {
                if (!Array.isArray(object.features))
                    throw TypeError(".tutorial.Robot.features: array expected");
                message.features = [];
                for (let i = 0; i < object.features.length; ++i) {
                    if (typeof object.features[i] !== "object")
                        throw TypeError(".tutorial.Robot.features: object expected");
                    message.features[i] = $root.tutorial.Robot.Feature.fromObject(object.features[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a Robot message. Also converts values to other types if specified.
         * @function toObject
         * @memberof tutorial.Robot
         * @static
         * @param {tutorial.Robot} message Robot
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Robot.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.features = [];
            if (options.defaults) {
                object.name = "";
                object.id = 0;
            }
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.features && message.features.length) {
                object.features = [];
                for (let j = 0; j < message.features.length; ++j)
                    object.features[j] = $root.tutorial.Robot.Feature.toObject(message.features[j], options);
            }
            return object;
        };

        /**
         * Converts this Robot to JSON.
         * @function toJSON
         * @memberof tutorial.Robot
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Robot.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Robot
         * @function getTypeUrl
         * @memberof tutorial.Robot
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Robot.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/tutorial.Robot";
        };

        Robot.Feature = (function() {

            /**
             * Properties of a Feature.
             * @memberof tutorial.Robot
             * @interface IFeature
             * @property {string|null} [name] Feature name
             * @property {string|null} [description] Feature description
             */

            /**
             * Constructs a new Feature.
             * @memberof tutorial.Robot
             * @classdesc Represents a Feature.
             * @implements IFeature
             * @constructor
             * @param {tutorial.Robot.IFeature=} [properties] Properties to set
             */
            function Feature(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Feature name.
             * @member {string} name
             * @memberof tutorial.Robot.Feature
             * @instance
             */
            Feature.prototype.name = "";

            /**
             * Feature description.
             * @member {string} description
             * @memberof tutorial.Robot.Feature
             * @instance
             */
            Feature.prototype.description = "";

            /**
             * Creates a new Feature instance using the specified properties.
             * @function create
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {tutorial.Robot.IFeature=} [properties] Properties to set
             * @returns {tutorial.Robot.Feature} Feature instance
             */
            Feature.create = function create(properties) {
                return new Feature(properties);
            };

            /**
             * Encodes the specified Feature message. Does not implicitly {@link tutorial.Robot.Feature.verify|verify} messages.
             * @function encode
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {tutorial.Robot.IFeature} message Feature message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Feature.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.description != null && Object.hasOwnProperty.call(message, "description"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.description);
                return writer;
            };

            /**
             * Encodes the specified Feature message, length delimited. Does not implicitly {@link tutorial.Robot.Feature.verify|verify} messages.
             * @function encodeDelimited
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {tutorial.Robot.IFeature} message Feature message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Feature.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Feature message from the specified reader or buffer.
             * @function decode
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {tutorial.Robot.Feature} Feature
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Feature.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.tutorial.Robot.Feature();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1: {
                            message.name = reader.string();
                            break;
                        }
                    case 2: {
                            message.description = reader.string();
                            break;
                        }
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Feature message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {tutorial.Robot.Feature} Feature
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Feature.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Feature message.
             * @function verify
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Feature.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.description != null && message.hasOwnProperty("description"))
                    if (!$util.isString(message.description))
                        return "description: string expected";
                return null;
            };

            /**
             * Creates a Feature message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {tutorial.Robot.Feature} Feature
             */
            Feature.fromObject = function fromObject(object) {
                if (object instanceof $root.tutorial.Robot.Feature)
                    return object;
                let message = new $root.tutorial.Robot.Feature();
                if (object.name != null)
                    message.name = String(object.name);
                if (object.description != null)
                    message.description = String(object.description);
                return message;
            };

            /**
             * Creates a plain object from a Feature message. Also converts values to other types if specified.
             * @function toObject
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {tutorial.Robot.Feature} message Feature
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Feature.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                let object = {};
                if (options.defaults) {
                    object.name = "";
                    object.description = "";
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.description != null && message.hasOwnProperty("description"))
                    object.description = message.description;
                return object;
            };

            /**
             * Converts this Feature to JSON.
             * @function toJSON
             * @memberof tutorial.Robot.Feature
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Feature.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Gets the default type url for Feature
             * @function getTypeUrl
             * @memberof tutorial.Robot.Feature
             * @static
             * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns {string} The default type url
             */
            Feature.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
                if (typeUrlPrefix === undefined) {
                    typeUrlPrefix = "type.googleapis.com";
                }
                return typeUrlPrefix + "/tutorial.Robot.Feature";
            };

            return Feature;
        })();

        return Robot;
    })();

    tutorial.Task = (function() {

        /**
         * Properties of a Task.
         * @memberof tutorial
         * @interface ITask
         * @property {string|null} [title] Task title
         * @property {string|null} [description] Task description
         * @property {google.protobuf.ITimestamp|null} [dueDate] Task dueDate
         * @property {google.protobuf.IAny|null} [doneBy] Task doneBy
         */

        /**
         * Constructs a new Task.
         * @memberof tutorial
         * @classdesc Represents a Task.
         * @implements ITask
         * @constructor
         * @param {tutorial.ITask=} [properties] Properties to set
         */
        function Task(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Task title.
         * @member {string} title
         * @memberof tutorial.Task
         * @instance
         */
        Task.prototype.title = "";

        /**
         * Task description.
         * @member {string} description
         * @memberof tutorial.Task
         * @instance
         */
        Task.prototype.description = "";

        /**
         * Task dueDate.
         * @member {google.protobuf.ITimestamp|null|undefined} dueDate
         * @memberof tutorial.Task
         * @instance
         */
        Task.prototype.dueDate = null;

        /**
         * Task doneBy.
         * @member {google.protobuf.IAny|null|undefined} doneBy
         * @memberof tutorial.Task
         * @instance
         */
        Task.prototype.doneBy = null;

        /**
         * Creates a new Task instance using the specified properties.
         * @function create
         * @memberof tutorial.Task
         * @static
         * @param {tutorial.ITask=} [properties] Properties to set
         * @returns {tutorial.Task} Task instance
         */
        Task.create = function create(properties) {
            return new Task(properties);
        };

        /**
         * Encodes the specified Task message. Does not implicitly {@link tutorial.Task.verify|verify} messages.
         * @function encode
         * @memberof tutorial.Task
         * @static
         * @param {tutorial.ITask} message Task message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Task.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.title != null && Object.hasOwnProperty.call(message, "title"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.title);
            if (message.description != null && Object.hasOwnProperty.call(message, "description"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.description);
            if (message.dueDate != null && Object.hasOwnProperty.call(message, "dueDate"))
                $root.google.protobuf.Timestamp.encode(message.dueDate, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.doneBy != null && Object.hasOwnProperty.call(message, "doneBy"))
                $root.google.protobuf.Any.encode(message.doneBy, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified Task message, length delimited. Does not implicitly {@link tutorial.Task.verify|verify} messages.
         * @function encodeDelimited
         * @memberof tutorial.Task
         * @static
         * @param {tutorial.ITask} message Task message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Task.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Task message from the specified reader or buffer.
         * @function decode
         * @memberof tutorial.Task
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {tutorial.Task} Task
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Task.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.tutorial.Task();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.title = reader.string();
                        break;
                    }
                case 2: {
                        message.description = reader.string();
                        break;
                    }
                case 3: {
                        message.dueDate = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    }
                case 4: {
                        message.doneBy = $root.google.protobuf.Any.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Task message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof tutorial.Task
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {tutorial.Task} Task
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Task.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Task message.
         * @function verify
         * @memberof tutorial.Task
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Task.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.title != null && message.hasOwnProperty("title"))
                if (!$util.isString(message.title))
                    return "title: string expected";
            if (message.description != null && message.hasOwnProperty("description"))
                if (!$util.isString(message.description))
                    return "description: string expected";
            if (message.dueDate != null && message.hasOwnProperty("dueDate")) {
                let error = $root.google.protobuf.Timestamp.verify(message.dueDate);
                if (error)
                    return "dueDate." + error;
            }
            if (message.doneBy != null && message.hasOwnProperty("doneBy")) {
                let error = $root.google.protobuf.Any.verify(message.doneBy);
                if (error)
                    return "doneBy." + error;
            }
            return null;
        };

        /**
         * Creates a Task message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof tutorial.Task
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {tutorial.Task} Task
         */
        Task.fromObject = function fromObject(object) {
            if (object instanceof $root.tutorial.Task)
                return object;
            let message = new $root.tutorial.Task();
            if (object.title != null)
                message.title = String(object.title);
            if (object.description != null)
                message.description = String(object.description);
            if (object.dueDate != null) {
                if (typeof object.dueDate !== "object")
                    throw TypeError(".tutorial.Task.dueDate: object expected");
                message.dueDate = $root.google.protobuf.Timestamp.fromObject(object.dueDate);
            }
            if (object.doneBy != null) {
                if (typeof object.doneBy !== "object")
                    throw TypeError(".tutorial.Task.doneBy: object expected");
                message.doneBy = $root.google.protobuf.Any.fromObject(object.doneBy);
            }
            return message;
        };

        /**
         * Creates a plain object from a Task message. Also converts values to other types if specified.
         * @function toObject
         * @memberof tutorial.Task
         * @static
         * @param {tutorial.Task} message Task
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Task.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.title = "";
                object.description = "";
                object.dueDate = null;
                object.doneBy = null;
            }
            if (message.title != null && message.hasOwnProperty("title"))
                object.title = message.title;
            if (message.description != null && message.hasOwnProperty("description"))
                object.description = message.description;
            if (message.dueDate != null && message.hasOwnProperty("dueDate"))
                object.dueDate = $root.google.protobuf.Timestamp.toObject(message.dueDate, options);
            if (message.doneBy != null && message.hasOwnProperty("doneBy"))
                object.doneBy = $root.google.protobuf.Any.toObject(message.doneBy, options);
            return object;
        };

        /**
         * Converts this Task to JSON.
         * @function toJSON
         * @memberof tutorial.Task
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Task.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Task
         * @function getTypeUrl
         * @memberof tutorial.Task
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Task.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/tutorial.Task";
        };

        return Task;
    })();

    return tutorial;
})();

export const google = $root.google = (() => {

    /**
     * Namespace google.
     * @exports google
     * @namespace
     */
    const google = {};

    google.protobuf = (function() {

        /**
         * Namespace protobuf.
         * @memberof google
         * @namespace
         */
        const protobuf = {};

        protobuf.Timestamp = (function() {

            /**
             * Properties of a Timestamp.
             * @memberof google.protobuf
             * @interface ITimestamp
             * @property {number|Long|null} [seconds] Timestamp seconds
             * @property {number|null} [nanos] Timestamp nanos
             */

            /**
             * Constructs a new Timestamp.
             * @memberof google.protobuf
             * @classdesc Represents a Timestamp.
             * @implements ITimestamp
             * @constructor
             * @param {google.protobuf.ITimestamp=} [properties] Properties to set
             */
            function Timestamp(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Timestamp seconds.
             * @member {number|Long} seconds
             * @memberof google.protobuf.Timestamp
             * @instance
             */
            Timestamp.prototype.seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

            /**
             * Timestamp nanos.
             * @member {number} nanos
             * @memberof google.protobuf.Timestamp
             * @instance
             */
            Timestamp.prototype.nanos = 0;

            /**
             * Creates a new Timestamp instance using the specified properties.
             * @function create
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.ITimestamp=} [properties] Properties to set
             * @returns {google.protobuf.Timestamp} Timestamp instance
             */
            Timestamp.create = function create(properties) {
                return new Timestamp(properties);
            };

            /**
             * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.ITimestamp} message Timestamp message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Timestamp.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.seconds != null && Object.hasOwnProperty.call(message, "seconds"))
                    writer.uint32(/* id 1, wireType 0 =*/8).int64(message.seconds);
                if (message.nanos != null && Object.hasOwnProperty.call(message, "nanos"))
                    writer.uint32(/* id 2, wireType 0 =*/16).int32(message.nanos);
                return writer;
            };

            /**
             * Encodes the specified Timestamp message, length delimited. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @function encodeDelimited
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.ITimestamp} message Timestamp message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Timestamp.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Timestamp message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.Timestamp} Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Timestamp.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Timestamp();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1: {
                            message.seconds = reader.int64();
                            break;
                        }
                    case 2: {
                            message.nanos = reader.int32();
                            break;
                        }
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Timestamp message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {google.protobuf.Timestamp} Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Timestamp.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Timestamp message.
             * @function verify
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Timestamp.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.seconds != null && message.hasOwnProperty("seconds"))
                    if (!$util.isInteger(message.seconds) && !(message.seconds && $util.isInteger(message.seconds.low) && $util.isInteger(message.seconds.high)))
                        return "seconds: integer|Long expected";
                if (message.nanos != null && message.hasOwnProperty("nanos"))
                    if (!$util.isInteger(message.nanos))
                        return "nanos: integer expected";
                return null;
            };

            /**
             * Creates a Timestamp message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {google.protobuf.Timestamp} Timestamp
             */
            Timestamp.fromObject = function fromObject(object) {
                if (object instanceof $root.google.protobuf.Timestamp)
                    return object;
                let message = new $root.google.protobuf.Timestamp();
                if (object.seconds != null)
                    if ($util.Long)
                        (message.seconds = $util.Long.fromValue(object.seconds)).unsigned = false;
                    else if (typeof object.seconds === "string")
                        message.seconds = parseInt(object.seconds, 10);
                    else if (typeof object.seconds === "number")
                        message.seconds = object.seconds;
                    else if (typeof object.seconds === "object")
                        message.seconds = new $util.LongBits(object.seconds.low >>> 0, object.seconds.high >>> 0).toNumber();
                if (object.nanos != null)
                    message.nanos = object.nanos | 0;
                return message;
            };

            /**
             * Creates a plain object from a Timestamp message. Also converts values to other types if specified.
             * @function toObject
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.Timestamp} message Timestamp
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Timestamp.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                let object = {};
                if (options.defaults) {
                    if ($util.Long) {
                        let long = new $util.Long(0, 0, false);
                        object.seconds = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                    } else
                        object.seconds = options.longs === String ? "0" : 0;
                    object.nanos = 0;
                }
                if (message.seconds != null && message.hasOwnProperty("seconds"))
                    if (typeof message.seconds === "number")
                        object.seconds = options.longs === String ? String(message.seconds) : message.seconds;
                    else
                        object.seconds = options.longs === String ? $util.Long.prototype.toString.call(message.seconds) : options.longs === Number ? new $util.LongBits(message.seconds.low >>> 0, message.seconds.high >>> 0).toNumber() : message.seconds;
                if (message.nanos != null && message.hasOwnProperty("nanos"))
                    object.nanos = message.nanos;
                return object;
            };

            /**
             * Converts this Timestamp to JSON.
             * @function toJSON
             * @memberof google.protobuf.Timestamp
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Timestamp.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Gets the default type url for Timestamp
             * @function getTypeUrl
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns {string} The default type url
             */
            Timestamp.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
                if (typeUrlPrefix === undefined) {
                    typeUrlPrefix = "type.googleapis.com";
                }
                return typeUrlPrefix + "/google.protobuf.Timestamp";
            };

            return Timestamp;
        })();

        protobuf.Any = (function() {

            /**
             * Properties of an Any.
             * @memberof google.protobuf
             * @interface IAny
             * @property {string|null} [type_url] Any type_url
             * @property {Uint8Array|null} [value] Any value
             */

            /**
             * Constructs a new Any.
             * @memberof google.protobuf
             * @classdesc Represents an Any.
             * @implements IAny
             * @constructor
             * @param {google.protobuf.IAny=} [properties] Properties to set
             */
            function Any(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Any type_url.
             * @member {string} type_url
             * @memberof google.protobuf.Any
             * @instance
             */
            Any.prototype.type_url = "";

            /**
             * Any value.
             * @member {Uint8Array} value
             * @memberof google.protobuf.Any
             * @instance
             */
            Any.prototype.value = $util.newBuffer([]);

            /**
             * Creates a new Any instance using the specified properties.
             * @function create
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.IAny=} [properties] Properties to set
             * @returns {google.protobuf.Any} Any instance
             */
            Any.create = function create(properties) {
                return new Any(properties);
            };

            /**
             * Encodes the specified Any message. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.IAny} message Any message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Any.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.type_url != null && Object.hasOwnProperty.call(message, "type_url"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.type_url);
                if (message.value != null && Object.hasOwnProperty.call(message, "value"))
                    writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.value);
                return writer;
            };

            /**
             * Encodes the specified Any message, length delimited. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
             * @function encodeDelimited
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.IAny} message Any message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Any.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes an Any message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.Any
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.Any} Any
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Any.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Any();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1: {
                            message.type_url = reader.string();
                            break;
                        }
                    case 2: {
                            message.value = reader.bytes();
                            break;
                        }
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes an Any message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof google.protobuf.Any
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {google.protobuf.Any} Any
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Any.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies an Any message.
             * @function verify
             * @memberof google.protobuf.Any
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Any.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.type_url != null && message.hasOwnProperty("type_url"))
                    if (!$util.isString(message.type_url))
                        return "type_url: string expected";
                if (message.value != null && message.hasOwnProperty("value"))
                    if (!(message.value && typeof message.value.length === "number" || $util.isString(message.value)))
                        return "value: buffer expected";
                return null;
            };

            /**
             * Creates an Any message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof google.protobuf.Any
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {google.protobuf.Any} Any
             */
            Any.fromObject = function fromObject(object) {
                if (object instanceof $root.google.protobuf.Any)
                    return object;
                let message = new $root.google.protobuf.Any();
                if (object.type_url != null)
                    message.type_url = String(object.type_url);
                if (object.value != null)
                    if (typeof object.value === "string")
                        $util.base64.decode(object.value, message.value = $util.newBuffer($util.base64.length(object.value)), 0);
                    else if (object.value.length >= 0)
                        message.value = object.value;
                return message;
            };

            /**
             * Creates a plain object from an Any message. Also converts values to other types if specified.
             * @function toObject
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.Any} message Any
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Any.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                let object = {};
                if (options.defaults) {
                    object.type_url = "";
                    if (options.bytes === String)
                        object.value = "";
                    else {
                        object.value = [];
                        if (options.bytes !== Array)
                            object.value = $util.newBuffer(object.value);
                    }
                }
                if (message.type_url != null && message.hasOwnProperty("type_url"))
                    object.type_url = message.type_url;
                if (message.value != null && message.hasOwnProperty("value"))
                    object.value = options.bytes === String ? $util.base64.encode(message.value, 0, message.value.length) : options.bytes === Array ? Array.prototype.slice.call(message.value) : message.value;
                return object;
            };

            /**
             * Converts this Any to JSON.
             * @function toJSON
             * @memberof google.protobuf.Any
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Any.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Gets the default type url for Any
             * @function getTypeUrl
             * @memberof google.protobuf.Any
             * @static
             * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns {string} The default type url
             */
            Any.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
                if (typeUrlPrefix === undefined) {
                    typeUrlPrefix = "type.googleapis.com";
                }
                return typeUrlPrefix + "/google.protobuf.Any";
            };

            return Any;
        })();

        return protobuf;
    })();

    return google;
})();

export { $root as default };
