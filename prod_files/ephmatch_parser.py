# NOTE: you NEED PyYAML to use this!
import argparse, yaml, re, datetime as dt
from pathlib import Path

# under normal circumstances, this file should run once a year.

### parse arguments
parser = argparse.ArgumentParser()
parser.add_argument("file", help="the configuration file for wso-backend")
args = parser.parse_args()
file_path = Path(args.file)

### get data
if file_path.exists() and not file_path.is_dir():
    with open(file_path, "r") as x:
        data_raw = x.read()
        print(
            "[ephmatch_parser] successfully parsed at time: "
            + dt.datetime.now().isoformat()
        )
else:
    print("[ephmatch_parser] update failed at time: " + dt.datetime.now().isoformat())
    raise FileNotFoundError("file " + str(args.file) + " does not exist.")

### find the dates
year = dt.datetime.now().year

dates = {
    # winter study (usually)
    "date_winterstudy": {
        "start": dt.datetime(year, 1, 5).isoformat(),
        "end": dt.datetime(year, 2, 1).isoformat(),
    },
    # give everyone more time
    "date_prevalentines": {
        "start": dt.datetime(year, 2, 2).isoformat(),
        "end": dt.datetime(year, 2, 15).isoformat(),
    },
    # for the seniors
    "date_postvalentines": {
        "start": dt.datetime(year, 2, 16).isoformat(),
        "end": dt.datetime(year, 2, (14 + 7)).isoformat(),
        "senior_only": True,
    },
    # mid-march
    "date_march": {
        "start": dt.datetime(year, 3, 7).isoformat(),
        "end": dt.datetime(year, 3, (14 + 7)).isoformat(),
    },
    # first fortnight of april (so right after spring break most years)
    "date_april": {
        "start": dt.datetime(year, 4, 1).isoformat(),
        "end": dt.datetime(year, 4, (7 + 7)).isoformat(),
    },
    # the graduation season
    "date_graduation": {
        "start": dt.datetime(year, 5, 24).isoformat(),
        "end": dt.datetime(year, 7, 1).isoformat(),
        "senior_only": True,
    },
    # random stretch of summer
    "date_summer": {
        "start": dt.datetime(year, 7, 20).isoformat(),
        "end": dt.datetime(year, 8, 10).isoformat(),
        "senior_only": True,
    },
    # the mountain day season and halloween
    "date_mountainday": {
        "start": dt.datetime(year, 10, 4).isoformat(),
        "end": dt.datetime(year, 10, 31).isoformat(),
    },
    # the thanksgiving season (all possible dates)
    "date_thanksgiving": {
        "start": dt.datetime(year, 11, 21).isoformat(),
        "end": dt.datetime(year, 11, 30).isoformat(),
        "senior_only": True,
    },
    # around the holidays
    "date_holidays": {
        "start": dt.datetime(year, 12, 1).isoformat(),
        "end": dt.datetime(year, 12, 31).isoformat(),
    },
}

### stuff them into YAML format:
storage = []
for key in dates:
    storage += [dates[key]]

# then convert to yaml
storage = yaml.safe_dump(storage)
# they all start with two spaces
ephmatch_eras = "\n".join("  " + line for line in storage.splitlines())

### write back to the file
data_rawvalue = ephmatch_eras
# find all lines until the first with a non-whitespace character after our key
pattern = r"(?<=ephmatch_eras:)(.*?)(?=\n^\S)"
# the flags are important since this is a multiline regex
updated_block = re.sub(
    pattern, r"\n" + data_rawvalue, data_raw, flags=re.DOTALL | re.MULTILINE
)
with open(file_path, "w") as x:
    x.write(updated_block)
print("[ephmatch_parser] update complete at time: " + dt.datetime.now().isoformat())
