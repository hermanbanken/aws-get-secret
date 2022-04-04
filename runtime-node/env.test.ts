import { findAll } from "./env";
findAll({
    "FOO": "aws:///test",
    "FOO2": "aws:///test?template=notsupported",
});