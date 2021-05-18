import os
import sys

import yaml
from pydantic import BaseModel
import pydantic

from .Utils import is_frozen, get_exe_metadata

class Metadata(BaseModel, extra=pydantic.Extra.ignore):
    Version: str = "Development"
    CompanyName: str
    FileDescription: str
    OriginalFilename: str
    ProductName: str

    def __init__(self):
        if not is_frozen():
            with open('metadata.yml') as f:
                metadata_dict = yaml.load(f, Loader=yaml.FullLoader)

            super().__init__(**metadata_dict)

        else:
            data = get_exe_metadata(sys.argv[0])['StringFileInfo']
            data['Version'] = ".".join(data['ProductVersion'].split('.')[:3])

            super().__init__(**data)

metadata = Metadata()
