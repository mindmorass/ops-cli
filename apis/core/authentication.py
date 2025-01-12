import subprocess
from abc import ABC, abstractmethod
from typing import Dict, Optional


class Authenticator(ABC):
    @abstractmethod
    def authenticated(self) -> bool:
        """Check if authenticated with service"""
        pass


class AWSAuthenticator(Authenticator):
    def __init__(self, credentials: Optional[Dict] = None):
        """
        AWS Authentication using boto3 default chain if no credentials provided
        Args:
            credentials: Optional AWS credentials dict
        """
        self.credentials = credentials

    def authenticated(self) -> bool:
        try:
            # Use AWS CLI to validate authentication
            # Will use default credential chain if no credentials provided
            result = subprocess.run(
                ["aws", "sts", "get-caller-identity"],
                check=True,
                capture_output=True,
                text=True,
            )
            return True
        except subprocess.CalledProcessError:
            return False


class AzureAuthenticator(Authenticator):
    def __init__(self, credentials: Optional[Dict] = None):
        """
        Azure Authentication using Azure CLI default auth if no credentials provided
        Args:
            credentials: Optional Azure credentials dict
        """
        self.credentials = credentials

    def authenticated(self) -> bool:
        try:
            # Use Azure CLI to validate authentication
            # Will use CLI login if no credentials provided
            result = subprocess.run(
                ["az", "account", "show"],
                check=True,
                capture_output=True,
                text=True,
            )
            return True
        except subprocess.CalledProcessError:
            return False


class GCPAuthenticator(Authenticator):
    def __init__(self, credentials: Optional[Dict] = None):
        """
        GCP Authentication using gcloud default auth if no credentials provided
        Args:
            credentials: Optional GCP credentials dict
        """
        self.credentials = credentials

    def authenticated(self) -> bool:
        try:
            # Use gcloud CLI to validate authentication
            # Will use application default credentials if no credentials provided
            result = subprocess.run(
                ["gcloud", "auth", "print-identity-token"],
                check=True,
                capture_output=True,
                text=True,
            )
            return True
        except subprocess.CalledProcessError:
            return False


class AuthenticationApi:
    def __init__(self, auth_type: str, credentials: Optional[Dict] = None):
        """
        Initialize authentication for cloud services
        Args:
            auth_type: Authentication type ('aws', 'azure', 'gcp')
            credentials: Optional credentials dict (uses SDK defaults if None)
        """
        self.authenticator = self._get_authenticator(auth_type, credentials)

    def _get_authenticator(
        self, auth_type: str, credentials: Optional[Dict]
    ) -> Authenticator:
        """Get appropriate authenticator based on type"""
        authenticators = {
            "aws": AWSAuthenticator,
            "azure": AzureAuthenticator,
            "gcp": GCPAuthenticator,
        }

        authenticator_class = authenticators.get(auth_type.lower())
        if not authenticator_class:
            raise ValueError(f"Unsupported authentication type: {auth_type}")

        return authenticator_class(credentials)

    def is_authenticated(self) -> bool:
        """Check if authenticated with service"""
        return self.authenticator.authenticated()
