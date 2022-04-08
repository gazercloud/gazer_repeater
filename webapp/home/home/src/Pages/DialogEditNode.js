import React from 'react';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import Request from "../request";
import IconButton from "@material-ui/core/IconButton";
import EditIcon from '@material-ui/icons/Edit';
import Zoom from "@material-ui/core/Zoom";
import {Tooltip} from "@material-ui/core";

export default function DialogEditNode(props) {
    const [open, setOpen] = React.useState(false);
    const [text, setText] = React.useState("");
    const [error, setError] = React.useState("");
    const [nodes, setNodes] = React.useState({})

    const updateNode = () => {
        let req = {
            node_id: props.NodeId,
            name: text
        }
        Request('s-node-update', req)
            .then((res) => {
                if (res.status === 200) {
                    res.json().then(
                        (result) => {
                            handleClose()
                            props.OnNeedUpdate()
                        }
                    );
                } else {
                    res.json().then(
                        (result) => {
                            setError(result.error)
                        }
                    );
                }
            }).catch((e) => {
            setError(e.message)
        });
    }

    const requestNodes = () => {
        let req = {
        }
        Request('s-registered-nodes', req)
            .then((res) => {
                if (res.status === 200) {
                    res.json().then(
                        (result) => {
                            setNodes(result)
                            if (result !== undefined && result.items !== undefined) {
                                for (let i = 0; i < result.items.length; i++) {
                                    if (result.items[i].id === props.NodeId) {
                                        setText(result.items[i].name)
                                    }
                                }
                            }
                        }
                    );
                } else {
                    res.json().then(
                        (result) => {
                            setError(result.error)
                        }
                    );
                }
            }).catch((e) => {
            setError(e.message)
        });
    }


    const handleClickOpen = () => {
        setOpen(true);
        requestNodes();
    };

    const handleClose = () => {
        setOpen(false);
    };

    const handleOK = () => {
        updateNode();
    };

    const handleChange = (ev) => {
        setText(ev.target.value)
    };

    return (
        <div>
            <Tooltip title="Edit" TransitionComponent={Zoom}><IconButton><EditIcon fontSize="large" onClick={handleClickOpen} /></IconButton></Tooltip>
            <Dialog open={open} onClose={handleClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Edit node</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Edit node: {props.NodeId}
                    </DialogContentText>
                    <TextField
                        autoFocus
                        margin="dense"
                        id="name"
                        label="Name"
                        type="text"
                        value={text}
                        fullWidth
                        onChange={handleChange}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={handleOK} color="primary">
                        OK
                    </Button>
                </DialogActions>
            </Dialog>

            <div>{error}</div>
        </div>
    );
}
