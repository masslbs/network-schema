import * as $protobuf from "protobufjs";
import Long = require("long");
/** Namespace tutorial. */
export namespace tutorial {

    /** Properties of a Person. */
    interface IPerson {

        /** Person name */
        name?: (string|null);

        /** Person id */
        id?: (number|null);

        /** Person age */
        age?: (number|null);
    }

    /** Represents a Person. */
    class Person implements IPerson {

        /**
         * Constructs a new Person.
         * @param [properties] Properties to set
         */
        constructor(properties?: tutorial.IPerson);

        /** Person name. */
        public name: string;

        /** Person id. */
        public id: number;

        /** Person age. */
        public age: number;

        /**
         * Creates a new Person instance using the specified properties.
         * @param [properties] Properties to set
         * @returns Person instance
         */
        public static create(properties?: tutorial.IPerson): tutorial.Person;

        /**
         * Encodes the specified Person message. Does not implicitly {@link tutorial.Person.verify|verify} messages.
         * @param message Person message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: tutorial.IPerson, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified Person message, length delimited. Does not implicitly {@link tutorial.Person.verify|verify} messages.
         * @param message Person message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: tutorial.IPerson, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a Person message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Person
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): tutorial.Person;

        /**
         * Decodes a Person message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns Person
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): tutorial.Person;

        /**
         * Verifies a Person message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a Person message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns Person
         */
        public static fromObject(object: { [k: string]: any }): tutorial.Person;

        /**
         * Creates a plain object from a Person message. Also converts values to other types if specified.
         * @param message Person
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: tutorial.Person, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this Person to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for Person
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a Robot. */
    interface IRobot {

        /** Robot name */
        name?: (string|null);

        /** Robot id */
        id?: (number|null);

        /** Robot features */
        features?: (tutorial.Robot.IFeature[]|null);
    }

    /** Represents a Robot. */
    class Robot implements IRobot {

        /**
         * Constructs a new Robot.
         * @param [properties] Properties to set
         */
        constructor(properties?: tutorial.IRobot);

        /** Robot name. */
        public name: string;

        /** Robot id. */
        public id: number;

        /** Robot features. */
        public features: tutorial.Robot.IFeature[];

        /**
         * Creates a new Robot instance using the specified properties.
         * @param [properties] Properties to set
         * @returns Robot instance
         */
        public static create(properties?: tutorial.IRobot): tutorial.Robot;

        /**
         * Encodes the specified Robot message. Does not implicitly {@link tutorial.Robot.verify|verify} messages.
         * @param message Robot message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: tutorial.IRobot, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified Robot message, length delimited. Does not implicitly {@link tutorial.Robot.verify|verify} messages.
         * @param message Robot message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: tutorial.IRobot, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a Robot message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Robot
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): tutorial.Robot;

        /**
         * Decodes a Robot message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns Robot
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): tutorial.Robot;

        /**
         * Verifies a Robot message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a Robot message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns Robot
         */
        public static fromObject(object: { [k: string]: any }): tutorial.Robot;

        /**
         * Creates a plain object from a Robot message. Also converts values to other types if specified.
         * @param message Robot
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: tutorial.Robot, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this Robot to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for Robot
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    namespace Robot {

        /** Properties of a Feature. */
        interface IFeature {

            /** Feature name */
            name?: (string|null);

            /** Feature description */
            description?: (string|null);
        }

        /** Represents a Feature. */
        class Feature implements IFeature {

            /**
             * Constructs a new Feature.
             * @param [properties] Properties to set
             */
            constructor(properties?: tutorial.Robot.IFeature);

            /** Feature name. */
            public name: string;

            /** Feature description. */
            public description: string;

            /**
             * Creates a new Feature instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Feature instance
             */
            public static create(properties?: tutorial.Robot.IFeature): tutorial.Robot.Feature;

            /**
             * Encodes the specified Feature message. Does not implicitly {@link tutorial.Robot.Feature.verify|verify} messages.
             * @param message Feature message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: tutorial.Robot.IFeature, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Encodes the specified Feature message, length delimited. Does not implicitly {@link tutorial.Robot.Feature.verify|verify} messages.
             * @param message Feature message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encodeDelimited(message: tutorial.Robot.IFeature, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Feature message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Feature
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): tutorial.Robot.Feature;

            /**
             * Decodes a Feature message from the specified reader or buffer, length delimited.
             * @param reader Reader or buffer to decode from
             * @returns Feature
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): tutorial.Robot.Feature;

            /**
             * Verifies a Feature message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);

            /**
             * Creates a Feature message from a plain object. Also converts values to their respective internal types.
             * @param object Plain object
             * @returns Feature
             */
            public static fromObject(object: { [k: string]: any }): tutorial.Robot.Feature;

            /**
             * Creates a plain object from a Feature message. Also converts values to other types if specified.
             * @param message Feature
             * @param [options] Conversion options
             * @returns Plain object
             */
            public static toObject(message: tutorial.Robot.Feature, options?: $protobuf.IConversionOptions): { [k: string]: any };

            /**
             * Converts this Feature to JSON.
             * @returns JSON object
             */
            public toJSON(): { [k: string]: any };

            /**
             * Gets the default type url for Feature
             * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns The default type url
             */
            public static getTypeUrl(typeUrlPrefix?: string): string;
        }
    }

    /** Properties of a Task. */
    interface ITask {

        /** Task title */
        title?: (string|null);

        /** Task description */
        description?: (string|null);

        /** Task dueDate */
        dueDate?: (google.protobuf.ITimestamp|null);

        /** Task doneBy */
        doneBy?: (google.protobuf.IAny|null);
    }

    /** Represents a Task. */
    class Task implements ITask {

        /**
         * Constructs a new Task.
         * @param [properties] Properties to set
         */
        constructor(properties?: tutorial.ITask);

        /** Task title. */
        public title: string;

        /** Task description. */
        public description: string;

        /** Task dueDate. */
        public dueDate?: (google.protobuf.ITimestamp|null);

        /** Task doneBy. */
        public doneBy?: (google.protobuf.IAny|null);

        /**
         * Creates a new Task instance using the specified properties.
         * @param [properties] Properties to set
         * @returns Task instance
         */
        public static create(properties?: tutorial.ITask): tutorial.Task;

        /**
         * Encodes the specified Task message. Does not implicitly {@link tutorial.Task.verify|verify} messages.
         * @param message Task message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: tutorial.ITask, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified Task message, length delimited. Does not implicitly {@link tutorial.Task.verify|verify} messages.
         * @param message Task message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: tutorial.ITask, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a Task message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Task
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): tutorial.Task;

        /**
         * Decodes a Task message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns Task
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): tutorial.Task;

        /**
         * Verifies a Task message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a Task message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns Task
         */
        public static fromObject(object: { [k: string]: any }): tutorial.Task;

        /**
         * Creates a plain object from a Task message. Also converts values to other types if specified.
         * @param message Task
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: tutorial.Task, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this Task to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for Task
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }
}

/** Namespace google. */
export namespace google {

    /** Namespace protobuf. */
    namespace protobuf {

        /** Properties of a Timestamp. */
        interface ITimestamp {

            /** Timestamp seconds */
            seconds?: (number|Long|null);

            /** Timestamp nanos */
            nanos?: (number|null);
        }

        /** Represents a Timestamp. */
        class Timestamp implements ITimestamp {

            /**
             * Constructs a new Timestamp.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.ITimestamp);

            /** Timestamp seconds. */
            public seconds: (number|Long);

            /** Timestamp nanos. */
            public nanos: number;

            /**
             * Creates a new Timestamp instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Timestamp instance
             */
            public static create(properties?: google.protobuf.ITimestamp): google.protobuf.Timestamp;

            /**
             * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @param message Timestamp message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.ITimestamp, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Encodes the specified Timestamp message, length delimited. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @param message Timestamp message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encodeDelimited(message: google.protobuf.ITimestamp, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Timestamp message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.Timestamp;

            /**
             * Decodes a Timestamp message from the specified reader or buffer, length delimited.
             * @param reader Reader or buffer to decode from
             * @returns Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): google.protobuf.Timestamp;

            /**
             * Verifies a Timestamp message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);

            /**
             * Creates a Timestamp message from a plain object. Also converts values to their respective internal types.
             * @param object Plain object
             * @returns Timestamp
             */
            public static fromObject(object: { [k: string]: any }): google.protobuf.Timestamp;

            /**
             * Creates a plain object from a Timestamp message. Also converts values to other types if specified.
             * @param message Timestamp
             * @param [options] Conversion options
             * @returns Plain object
             */
            public static toObject(message: google.protobuf.Timestamp, options?: $protobuf.IConversionOptions): { [k: string]: any };

            /**
             * Converts this Timestamp to JSON.
             * @returns JSON object
             */
            public toJSON(): { [k: string]: any };

            /**
             * Gets the default type url for Timestamp
             * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns The default type url
             */
            public static getTypeUrl(typeUrlPrefix?: string): string;
        }

        /** Properties of an Any. */
        interface IAny {

            /** Any type_url */
            type_url?: (string|null);

            /** Any value */
            value?: (Uint8Array|null);
        }

        /** Represents an Any. */
        class Any implements IAny {

            /**
             * Constructs a new Any.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IAny);

            /** Any type_url. */
            public type_url: string;

            /** Any value. */
            public value: Uint8Array;

            /**
             * Creates a new Any instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Any instance
             */
            public static create(properties?: google.protobuf.IAny): google.protobuf.Any;

            /**
             * Encodes the specified Any message. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
             * @param message Any message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.IAny, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Encodes the specified Any message, length delimited. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
             * @param message Any message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encodeDelimited(message: google.protobuf.IAny, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Any message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Any
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.Any;

            /**
             * Decodes an Any message from the specified reader or buffer, length delimited.
             * @param reader Reader or buffer to decode from
             * @returns Any
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): google.protobuf.Any;

            /**
             * Verifies an Any message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);

            /**
             * Creates an Any message from a plain object. Also converts values to their respective internal types.
             * @param object Plain object
             * @returns Any
             */
            public static fromObject(object: { [k: string]: any }): google.protobuf.Any;

            /**
             * Creates a plain object from an Any message. Also converts values to other types if specified.
             * @param message Any
             * @param [options] Conversion options
             * @returns Plain object
             */
            public static toObject(message: google.protobuf.Any, options?: $protobuf.IConversionOptions): { [k: string]: any };

            /**
             * Converts this Any to JSON.
             * @returns JSON object
             */
            public toJSON(): { [k: string]: any };

            /**
             * Gets the default type url for Any
             * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns The default type url
             */
            public static getTypeUrl(typeUrlPrefix?: string): string;
        }
    }
}
