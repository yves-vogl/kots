import * as React from "react";
import { graphql, compose, withApollo } from "react-apollo";
import { withRouter } from "react-router-dom";
import Dropzone from "react-dropzone";
import isEmpty from "lodash/isEmpty";
import { uploadKotsLicense } from "../mutations/AppsMutations";

import "../scss/components/troubleshoot/UploadSupportBundleModal.scss";
import "../scss/components/Login.scss";

class UploadLicenseFile extends React.Component {
  state = {
    licenseFile: {},
    licenseValue: "",
    fileUploading: false
  }

  clearFile = () => {
    this.setState({ licenseFile: {}, licenseValue: "" });
  }

  uploadToS3 = async () => {
    const { licenseValue } = this.state;
    this.setState({ fileUploading: true });
    try {
      await this.props.uploadKotsLicense(licenseValue);
      this.setState({ fileUploading: false });
      this.props.history.replace("/watches");
    } catch (err) {
      this.setState({ fileUploading: false });
      console.log(err);
    }
  }

  onDrop = async (files) => {
    const content = await files[0].text();
    this.setState({
      licenseFile: files[0],
      licenseValue: content
    });
  }

  render() {
    const { licenseFile, fileUploading } = this.state;
    const hasFile = licenseFile && !isEmpty(licenseFile);

    return (
      <div className="UploadLicenseFile--wrapper container flex-column flex1 u-overflow--auto Login-wrapper justifyContent--center alignItems--center">
        <div className="LoginBox-wrapper u-flexTabletReflow flex-auto">
          <div className="flex-auto flex-column login-form-wrapper secure-console justifyContent--center">
            <div className="flex-column alignItems--center">
              <span className="icon ship-login-icon"></span>
              <p className="u-marginTop--10 u-paddingTop--5 u-fontSize--header u-color--tuna u-fontWeight--bold">Upload your license file</p>
            </div>
            <div className="u-marginTop--30 flex">
              <div className={`FileUpload-wrapper flex1 ${hasFile ? "has-file" : ""}`}>
                <Dropzone
                  className="Dropzone-wrapper"
                  accept={["application/x-yaml", ".yaml", ".yml"]}
                  onDropAccepted={this.onDrop}
                  multiple={false}
                >
                  {hasFile ?
                    <div className="has-file-wrapper">
                      <p className="u-fontSize--normal u-fontWeight--medium">{licenseFile.name}</p>
                    </div>
                    :
                    <div className="u-textAlign--center">
                      <p className="u-fontSize--normal u-color--tundora u-fontWeight--medium u-lineHeight--normal">Drag your license here or <span className="u-color--astral u-fontWeight--medium u-textDecoration--underlineOnHover">choose a file to upload</span></p>
                      <p className="u-fontSize--normal u-color--dustyGray u-fontWeight--normal u-lineHeight--normal u-marginTop--10">This will be a .yaml file your vendor provided. Contact them if you are unable to locate a license file.</p>
                    </div>
                  }
                </Dropzone>
              </div>
              {hasFile && 
                <div className="flex-auto flex-column u-marginLeft--10 justifyContent--center">
                  <button
                    type="button"
                    className="btn primary large flex-auto"
                    onClick={this.uploadToS3}
                    disabled={fileUploading || !hasFile}
                  >
                    {fileUploading ? "Uploading" : "Upload license"}
                  </button>
                </div>
              }
            </div>
            {hasFile &&
              <div className="u-marginTop--10">
                <span className="replicated-link u-fontSize--small" onClick={this.clearFile}>Select a different file</span>
              </div>
            }
          </div>
        </div>
      </div>
    );
  }
}

export default compose(
  withRouter,
  withApollo,
  graphql(uploadKotsLicense, {
    props: ({ mutate }) => ({
      uploadKotsLicense: (value) => mutate({ variables: { value } })
    })
  }),
)(UploadLicenseFile);
