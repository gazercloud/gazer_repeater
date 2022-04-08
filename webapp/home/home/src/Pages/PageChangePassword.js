import React, {useState} from 'react';
import {Button} from "@material-ui/core";
import Request from "../request";
import Container from "@material-ui/core/Container";
import CssBaseline from "@material-ui/core/CssBaseline";
import Avatar from "@material-ui/core/Avatar";
import LockOutlinedIcon from "@material-ui/icons/LockOutlined";
import Typography from "@material-ui/core/Typography";
import TextField from "@material-ui/core/TextField";
import {makeStyles} from "@material-ui/core/styles";

const useStyles = makeStyles((theme) => ({
    paper: {
        marginTop: theme.spacing(8),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    avatar: {
        margin: theme.spacing(1),
    },
    form: {
        width: '100%', // Fix IE 11 issue.
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
}));

export default function PageChangePassword(props) {
    const classes = useStyles();
    const [password1, setPassword1] = React.useState("")
    const [password2, setPassword2] = React.useState("")
    const [message, setMessage] = React.useState("")
    const [messageIsError, setMessageIsError] = React.useState(false)

    const requestChangePassword = () => {
        setMessageIsError(false)
        setMessage("processing ...")

        if (password1 !== password2) {
            setMessageIsError(true)
            setMessage("passwords do not match")
            return
        }

        let req = {
            "password": password1,
        }
        Request('s-change-password', req)
            .then((res) => {
                if (res.status === 200) {
                    res.text().then(
                        (result) => {
                            try {
                                setMessageIsError(false)
                                setPassword1("")
                                setPassword2("")
                                setMessage("Your password has been changed.")
                                props.OnNeedUpdate()
                            } catch (e) {
                                setMessage("wrong server response")
                            }

                        }
                    )
                    return
                }
                if (res.status === 500) {
                    res.json().then(
                        (result) => {
                            setMessageIsError(true)
                            setMessage("Error: " + result.error)
                        }
                    );
                    return
                }

                res.text().then(
                    (result) => {
                        setMessage("Error " + res.status + ": " + result)
                    }
                )
            });

    }

    const onClickOK = () => {
        requestChangePassword()
    }

    const [firstRendering, setFirstRendering] = useState(true)
    if (firstRendering) {
        setPassword1("")
        setPassword2("")
        setMessageIsError(false)
        setMessage("")
        props.OnTitleUpdate("Gazer.Cloud - Change password")
        setFirstRendering(false)
    }

    return (
        <Container component="main" maxWidth="xs">
            <CssBaseline />
            <div className={classes.paper}>
                <Avatar className={classes.avatar}>
                    <LockOutlinedIcon />
                </Avatar>
                <Typography component="h1" variant="h5">
                    Change password
                </Typography>
                <form className={classes.form} noValidate>
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        name="password1"
                        label="NEW PASSWORD"
                        type="password"
                        id="password1"
                        autoComplete="current-password"
                        value={password1}
                        onChange={(event) => {
                            setPassword1(event.target.value)
                        }}
                    />
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        name="password2"
                        label="CONFIRM NEW PASSWORD"
                        type="password"
                        id="password2"
                        autoComplete="current-password"
                        value={password2}
                        onChange={(event) => {
                            setPassword2(event.target.value)
                        }}
                    />
                    <Button
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                        onClick={onClickOK}
                    >
                        Apply
                    </Button>
                </form>
            </div>
            {messageIsError?
                <div style={{color:'#F20', fontSize: '24pt'}}>
                    {message}
                </div>:
                <div style={{color:'#082', fontSize: '24pt'}}>
                    {message}
                </div>
            }
        </Container>
    )
}
