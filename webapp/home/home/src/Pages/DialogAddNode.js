import React from 'react';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import Request from "../request";
import AddIcon from '@material-ui/icons/Add';
import IconButton from "@material-ui/core/IconButton";
import {Tooltip} from "@material-ui/core";
import Zoom from '@material-ui/core/Zoom';

export default function DialogAddNode(props) {
    const [open, setOpen] = React.useState(false);
    const [text, setText] = React.useState("");
    const [message, setMessage] = React.useState("")
    const [messageIsError, setMessageIsError] = React.useState(false)

    const addNode = () => {
        setMessageIsError(false)
        let req = {
            name: text
        }
        Request('s-node-add', req)
            .then((res) => {
                if (res.status === 200) {
                    res.json().then(
                        (result) => {
                            handleClose()
                            setMessage("")
                            setMessageIsError(false)
                            props.OnNeedUpdate()
                        }
                    );
                } else {
                    res.json().then(
                        (result) => {
                            setMessage(result.error)
                            setMessageIsError(true)
                        }
                    );
                }
            }).catch((e) => {
            setMessage(e.message)
            setMessageIsError(true)
        });
    }


    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
    };

    const handleOK = () => {
        addNode();
    };

    const handleChange = (ev) => {
        setText(ev.target.value)
    };

    return (
        <div>
            <Tooltip title="Add node" TransitionComponent={Zoom}>
                <Button variant="contained" color="primary" startIcon={<AddIcon fontSize="large"></AddIcon>} onClick={handleClickOpen}>ADD NODE</Button>
            </Tooltip>
            <Dialog open={open} onClose={handleClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Add node</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Add node
                    </DialogContentText>
                    <TextField
                        autoFocus
                        margin="dense"
                        id="name"
                        label="Name"
                        type="text"
                        value={text}
                        fullWidth
                        style={{minWidth: "300px"}}
                        onChange={handleChange}
                    />
                    <div>
                        {messageIsError?
                            <div style={{color:'#F20', fontSize: '24pt'}}>
                                {message}
                            </div>:
                            <div style={{color:'#082', fontSize: '24pt'}}>
                                {message}
                            </div>
                        }
                    </div>
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

        </div>
    );
}
