import React, {useState} from 'react';
import {Button, IconButton, LinearProgress, Link, Typography} from "@material-ui/core";
import Request from "../request";
import Grid from "@material-ui/core/Grid";
import CachedIcon from '@material-ui/icons/Cached';

function getWindow() {
    return window
}

function getCookie(name) {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? decodeURIComponent(matches[1]) : undefined;
}

export default function PageAccount(props) {
    const [accountInfo, setAccountInfo] = React.useState("")
    const [message, setMessage] = React.useState("")
    const [messageIsError, setMessageIsError] = React.useState(false)

    const [loadingProcessing, setLoadingProcessing] = React.useState(false)
    const [buyProcessing, setBuyProcessing] = React.useState(false)

    const requestAccountInfo = () => {
        if (loadingProcessing) {
            return
        }
        setLoadingProcessing(true)
        let req = {}
        Request('s-account-info', req)
            .then((res) => {
                if (res.status === 200) {
                    res.json().then(
                        (result) => {
                            setAccountInfo(result)
                            setMessageIsError(false)
                            setMessage("")
                            setLoadingProcessing(false)
                        }
                    );
                } else {
                    res.json().then(
                        (result) => {
                            setMessageIsError(true)
                            setMessage(result.error)
                            setLoadingProcessing(false)
                        }
                    );
                }
            }).catch(res => {
            setMessageIsError(true)
            setMessage(res.message)
            setLoadingProcessing(false)
        });
    }

    const requestBuy = (variant) => {
        if (buyProcessing) {
            return
        }
        setBuyProcessing(true)

        let req = {
            quantity: variant
        }
        Request('s-buy', req)
            .then((res) => {
                if (res.status === 200) {
                    res.json().then(
                        (result) => {
                            setBuyProcessing(false)
                            let win = getWindow().open(result.store_front_url, '_blank');
                            win.focus();
                        }
                    );
                } else {
                    res.json().then(
                        (result) => {
                            setMessageIsError(true)
                            setMessage(result.error)
                        }
                    );
                    setBuyProcessing(false)
                }
            }).catch(res => {
            setMessageIsError(true)
            setMessage(res.message)
            setBuyProcessing(false)
        });
    }

    const [firstRendering, setFirstRendering] = useState(true)
    if (firstRendering) {
        props.OnTitleUpdate("Gazer.Cloud - Account")
        requestAccountInfo()
        setFirstRendering(false)
    }

    const displayOffer = (name, comments, price, perNode1, perNode2, variantName, isMain) => {
        let minHeight = "350px"

        return (
            <div style={{
                margin: "10px",
                padding: "30px",
                border: "1px solid #2F528F",
                borderRadius: "5px",
                backgroundColor: "#222",
                minWidth: "300px",
                maxWidth: "300px"
            }}>
                <Grid container direction="column" alignItems="center" justify="center" style={{minHeight: minHeight}}>
                    <Grid item style={{fontSize: "24pt", color: "#FFD966"}}>
                        {name}
                    </Grid>
                    <Grid item style={{fontSize: "24pt", padding: "10px"}}>
                        <span style={{fontSize: "12pt"}}>$</span><span> {price}</span>
                    </Grid>
                    <Grid item style={{fontSize: "12pt", padding: "10px", textAlign: "center", color: "#888"}}>
                        <div>one-time payment</div>
                        <div>price without taxes</div>
                    </Grid>
                    <Grid item style={{verticalAlign: "sub", padding: "10px"}}>
                        <span style={{fontSize: "16pt", color: "#FFF"}}>${perNode1}</span>
                        <span style={{fontSize: "10pt", color: "#FFF"}}>.{perNode2}</span>
                        <span style={{fontSize: "10pt", color: "#777"}}> per node</span>
                    </Grid>
                    <Grid item style={{flexGrow: 1, fontSize: "16pt", color: "#777", textAlign: "center"}}>
                        {comments}
                    </Grid>
                    <Grid item style={{justifyContent: "center"}}>
                        <Button disabled={buyProcessing} variant="contained" color="primary" style={{margin: "10px", minWidth: "100px", fontSize: "24pt", color: "#FFF"}}
                                onClick={() => {
                                    requestBuy(variantName)
                                }}>BUY</Button>
                    </Grid>
                </Grid>
            </div>
        )
    }


    return (
        <Grid container direction="column" alignItems="center">
            <Grid item style={{alignSelf: "stretch"}}>
                        <div style={{borderRadius: "5px", padding: "10px", backgroundColor: "#222", margin: "10px"}}>
                            <Grid container alignItems="center">
                                <Grid item style={{flexGrow: 1}}>
                                    <div style={{fontSize: "14pt"}}>{accountInfo.email}</div>
                                </Grid>
                                <Grid item>
                                    <Button variant="outlined" color="primary" style={{margin: "10px"}} onClick={() => {
                                        props.OnNavigate("#form=change_password")
                                    }}>Change password</Button>
                                </Grid>
                                <Grid item>
                                    <Button variant="outlined" color="primary" style={{margin: "10px"}} onClick={() => {
                                        let req = {
                                            key: getCookie("session_token")
                                        }
                                        Request('session_close', req)
                                            .then((res) => {
                                                if (res.status === 200) {
                                                    props.OnNeedUpdate()
                                                } else {
                                                    props.OnNeedUpdate()
                                                }
                                            }).catch(res => {
                                            props.OnNeedUpdate()
                                        });
                                    }}>LogOut</Button>
                                </Grid>
                            </Grid>
                        </div>
            </Grid>
            <Grid item>
                <div style={{borderRadius: "5px", padding: "10px", backgroundColor: "#222", margin: "10px", maxWidth: "300px", minWidth: "300px"}}>
                    <Grid container direction="column" alignContent="center" alignItems="center">
                        <Grid item>
                            <div style={{textAlign: "center"}}>
                                <span style={{fontSize: "14pt", color: "#888"}}>{accountInfo.nodes_count} / </span>
                                <span style={{fontSize: "36pt", color: "#F80"}}>{accountInfo.max_nodes_count}</span>
                            </div>
                        </Grid>
                        <Grid item>
                            <div style={{fontSize: "16pt", textAlign: "center"}}>nodes</div>
                        </Grid>
                        <Grid item>
                            <IconButton style={{margin: "10px"}} onClick={() => {
                                requestAccountInfo()
                            }}><CachedIcon fontSize="large"/></IconButton>
                        </Grid>
                        <Grid item>
                            <LinearProgress hidden={!loadingProcessing} />
                        </Grid>
                    </Grid>
                </div>
            </Grid>
            <Grid item>
                {messageIsError ?
                    <div style={{color: '#F20', fontSize: '24pt'}}>
                        {message}
                    </div> :
                    <div style={{color: '#082', fontSize: '24pt'}}>
                        {message}
                    </div>
                }
            </Grid>
            <Grid item>
                <Grid container direction="row" alignItems="center" alignContent="center">
                    <Grid item>{displayOffer("5 nodes", " ", 24, 4, "80", 5, false)}</Grid>
                    <Grid item>{displayOffer("10 nodes", " ", 42, 4, "20", 10, true)}</Grid>
                    <Grid item>{displayOffer("50 nodes", " ", 75, 1, "50", 50, false)}</Grid>
                </Grid>
            </Grid>
            <Grid item>
                <LinearProgress hidden={!buyProcessing} />
            </Grid>
        </Grid>
    );
}
