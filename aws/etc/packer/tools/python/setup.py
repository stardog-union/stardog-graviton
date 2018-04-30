import os

import pip.req
import setuptools


requirements_path = os.path.join(os.path.dirname(__file__), "requirements.txt")
install_reqs = pip.req.parse_requirements(requirements_path)

setuptools.setup(name='stardog-cluster-utils',
      version="1.1",
      description="Tools for automating a stardog cluster",
      author="Stardog Union",
      url="http://www.stardog.com/",
      packages=setuptools.find_packages(),
      include_package_data=True,
      entry_points={
          'console_scripts': [
              "stardog-find-volume=stardog.cluster.find_volume:main",
              "stardog-wait-for-socket=stardog.cluster.wait_for_socket:main",
              "stardog-wait-for-pgm=stardog.cluster.test_program:main",
              "stardog-update=stardog.cluster.update_stardog:main",
              "stardog-refresh-binaries=stardog.cluster.refresh_stardog_binaries:main",
              "stardog-stop=stardog.cluster.stop_stardog:main",
              "stardog-start=stardog.cluster.start_stardog:main",
              "stardog-gather-logs=stardog.cluster.gather_log:main",
              "stardog-monitor-zk=stardog.cluster.monitor_zk:main",
              "stardog-jstack=stardog.cluster.run_jstack:main"
          ],
      },
      install_requires=["pyyaml == 3.10", "requests == 2.13.0"],

      package_data={"stardog.cluster": ["stardog/cluster/*"]},

      classifiers=[
          "Development Status :: 4 - Beta",
          "Environment :: Console",
          "Intended Audience :: System Administrators",
          "License :: OSI Approved :: Apache Software License",
          "Operating System :: POSIX :: Linux",
          "Programming Language :: Python",
          "Topic :: System :: Clustering",
          "Topic :: System :: Distributed Computing",
          "Programming Language :: Python :: 3 :: Only"
      ]
)
