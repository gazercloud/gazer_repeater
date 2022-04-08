import React from 'react';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import Request from "../request";
import DeleteIcon from '@material-ui/icons/Delete';
import IconButton from "@material-ui/core/IconButton";
import Zoom from "@material-ui/core/Zoom";
import {Tooltip} from "@material-ui/core";

export default function DialogRemoveNode(props) {
    const [open, setOpen] = React.useState(false);
    const [error, setError] = React.useState("");

    const removeNode = () => {
        let req = {
            node_id: props.NodeId
        }
        Request('s-node-remove', req)
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


    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
    };

    const handleOK = () => {
        removeNode();
    };

    return (
        <div>
            <Tooltip title="Remove" TransitionComponent={Zoom}><IconButton><DeleteIcon fontSize="large" onClick={handleClickOpen} /></IconButton></Tooltip>

            <Dialog open={open} onClose={handleClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Remove node</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Remove node: {props.NodeId}
                    </DialogContentText>
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
