import React from 'react';
import Typography from "@material-ui/core/Typography";

export default function PageAbout(props) {

    return (
        <div>
            <div style={{margin: "10px"}}><a style={{color: "#D27607", fontSize: "16pt"}} href="https://gazer.cloud/" target="_blank">Gazer.Cloud</a></div>
            <div style={{margin: "10px"}}><a style={{color: "#D27607", fontSize: "16pt"}} href="https://gazer.cloud/blog/how-to-install-node/" target="_blank">How to install a new node</a></div>
            <div style={{margin: "10px"}}><a style={{color: "#D27607", fontSize: "16pt"}} href="https://gazer.cloud/pricing/" target="_blank">Pricing</a></div>
        </div>
    );
}
