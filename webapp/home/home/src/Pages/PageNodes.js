import React, {useEffect, useState} from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Grid from "@material-ui/core/Grid";
import Request from "../request";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import BlurOnIcon from "@material-ui/icons/BlurOn";
import DialogAddNode from "./DialogAddNode";
import DialogEditNode from "./DialogEditNode";
import DialogRemoveNode from "./DialogRemoveNode";
import Zoom from "@material-ui/core/Zoom";
import IconButton from "@material-ui/core/IconButton";
import AddIcon from "@material-ui/icons/Add";
import {CircularProgress, Tooltip} from "@material-ui/core";
import RefreshIcon from '@material-ui/icons/Refresh';

const useStyles = makeStyles((theme) => ({
    root: {
        width: '100%',
        backgroundColor: theme.palette.background.paper,
    },
}));

function PageNodes(props) {
    const classes = useStyles();
    const [nodes, setNodes] = React.useState({})
    const [loading, setLoading] = React.useState(false)


    const btnStyle = (key) => {
        let borderTop = '0px solid #333333'
        let backColor = '#1E1E1E'
        let backColorHover = '#222222'

        if (currentItem === key) {
            return {
                margin: '10px',
                borderTop: borderTop,
                backgroundColor: backColor,
            }
        } else {
            if (hoverItem === key) {
                return {
                    padding: '20px',
                    minWidth: '300px',
                    maxWidth: '300px',
                    minHeight: '200px',
                    maxHeight: '200px',
                    borderRadius: '10px',
                    margin: '10px',
                    borderTop: borderTop,
                    backgroundColor: backColor,
                }
            } else {
                return {
                    padding: '20px',
                    minWidth: '300px',
                    maxWidth: '300px',
                    minHeight: '200px',
                    maxHeight: '200px',
                    borderRadius: '10px',
                    margin: '10px',
                    borderTop: borderTop,
                    backgroundColor: backColorHover,
                }
            }
        }
    }

    const [currentItem, setCurrentItem] = useState("")
    const [hoverItem, setHoverItem] = useState("")

    const btnClick = (item) => {
        let l = nodeUrl(item)
        var win = window.open(l, '_blank');
        win.focus();
    }

    const handleEnter = (ev, key) => {
        setHoverItem(ev)
    }

    const handleLeave = (ev, key) => {
        setHoverItem("")
    }

    const requestNodes = () => {
        setLoading(true)
        let req = {
        }
        Request('s-registered-nodes', req)
            .then((res) => {
                setLoading(false)
                if (res.status === 200) {
                    res.json().then(
                        (result) => {
                            setNodes(result)
                            props.OnTitleUpdate('Gazer - Nodes')
                        }
                    );
                } else {
                    res.json().then(
                        (result) => {
                            //setErrorMessage(result.error)
                        }
                    );
                }
            }).catch(() => {
            setLoading(false)
        });
    }

    const [firstRendering, setFirstRendering] = useState(true)
    if (firstRendering) {
        requestNodes()
        setFirstRendering(false)
    }

    const displayNodeValue = (unit) => {
        return (
            <div style={{
                fontSize: '24pt',
                color: '#080',
                overflow: 'hidden',
                height: '30pt',
                width: '260px',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
                textAlign: 'center'
            }}>
                {unit.name}
            </div>
        )
    }

    const nodeUrl = (item) => {
        return "https://client.gazer.cloud/"
    }

    const displayItem = (item) => {
        return (
            <Grid container direction='column'>
                <Grid item onClick={btnClick.bind(this, item)}  style={{cursor: 'pointer'}}>
                    <Grid container>
                        <Grid item>
                            <Grid container alignItems='center'>
                                {item.current_repeater !== ''?
                                    <BlurOnIcon color='primary' fontSize='large' style={{marginRight: '5px'}} /> :
                                    <BlurOnIcon color='disabled' fontSize='large' style={{marginRight: '5px'}} />
                                }

                                <Typography style={{
                                    fontSize: '14pt',
                                    overflow: 'hidden',
                                    height: '20pt',
                                    width: '200px',
                                    textOverflow: 'ellipsis',
                                    whiteSpace: 'nowrap',
                                }}>{item.id}</Typography>
                            </Grid>
                        </Grid>
                        <Grid item style={{marginTop: '10px'}}>
                            {displayNodeValue(item)}
                        </Grid>
                    </Grid>
                </Grid>
            </Grid>
        )
    }

    return (
        <div>
            <Grid container direction="row">
                <Grid item style={{padding: "10px"}}><DialogAddNode OnNeedUpdate={requestNodes} /></Grid>
                <Grid item style={{padding: "10px"}}>
                    <Tooltip title="Refresh" TransitionComponent={Zoom}>
                        <Button variant="outlined" color="primary" startIcon={<RefreshIcon fontSize="large"></RefreshIcon>} onClick={requestNodes.bind(this)}>Refresh</Button>
                    </Tooltip>
                </Grid>
                <Grid item style={{paddingTop: "16px"}}>
                    <a style={{color: "#D27607"}} href="https://gazer.cloud/blog/how-to-install-node/" target="_blank">How to install a new node</a>
                </Grid>
                <Grid item>
                    <div>{loading && <CircularProgress />}</div>
                </Grid>
            </Grid>
            <Grid container direction="column">
                <Grid item>
                    <Grid container direction="row">
                        {nodes !== undefined && nodes.items !== undefined ? nodes.items.map((item) => (
                            <Grid item
                                  key={"gazer-page_units-unit-" + item.id}
                                  button
                                  onMouseEnter={() => handleEnter(item.id)}
                                  onMouseLeave={() => handleLeave(item.id)}
                                  style={btnStyle(item.id)}
                            >
                                {displayItem(item)}
                                <Grid container direction="row" style={{marginTop: "30px"}}>
                                    <Grid item><DialogEditNode NodeId={item.id} OnNeedUpdate={requestNodes}/></Grid>
                                    <Grid item style={{flexGrow: "1"}}/>
                                    <Grid item><DialogRemoveNode NodeId={item.id} OnNeedUpdate={requestNodes}/></Grid>
                                </Grid>
                            </Grid>
                        )) : <div/>}
                    </Grid>
                </Grid>
            </Grid>
        </div>
    );
}

export default PageNodes;
